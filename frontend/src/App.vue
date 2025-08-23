<template>
  <div class="app-container">
    <!-- Application Toolbar -->
    <Toolbar
      :currentTheme="currentTheme"
      @toggleTheme="toggleTheme"
    />

    <!-- Fixed Header -->
    <header class="app-header">
      <h1 v-html="docHTMLTitle" class="document-title"></h1>
      <p v-html="docHTMLDate" class="document-dates"></p>
      <!-- Display error message if preset -->
      <p v-if="errorMessage" class="error-message">Error: {{ errorMessage }}</p>
    </header>

    <!-- Scrollable Main Content Area -->
    <main class="content-area">
      <article v-html="renderedHTML" id="content" class="markdown-body"></article>
    </main>
  </div>

  <!-- Help Dialog -->
  <Help
    ref="helpRef"
    :show="showHelpModal"
    @close="hideHelpModal"
  />

  <!-- Settings Dialog -->
  <Settings
    :show="showSettingsDialog"
    @close="hideSettingsDialog"
    @saved="onSettingsSaved"
  />
</template>

<script setup lang="ts">
import { ref, watch, nextTick,onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import { GetTheme, SetTheme } from '../wailsjs/go/main/App';
import Settings from './components/Settings.vue';
import Help from './components/Help.vue';
import Toolbar from './components/Toolbar.vue';
import mermaid from 'mermaid';

const renderedHTML = ref('<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>');
const docHTMLTitle = ref('');
const docHTMLDate = ref('');
const errorMessage = ref('');

// Modal reactive variables
const showHelpModal = ref(false);

// Settings dialog reactive variables
const showSettingsDialog = ref(false);

// Help component reference
const helpRef = ref<InstanceType<typeof Help>>();

const currentTheme = ref<'light' | 'dark'>('light');

// Function to add class to html and body elements
function addDocClass(thisClass: string) {
  document.documentElement.classList.add(thisClass);
  document.body.classList.add(thisClass);
}

// Function to remove class from html and body elements
function removeDocClass(thisClass: string) {
  document.documentElement.classList.remove(thisClass);
  document.body.classList.remove(thisClass);
}

// Function to toggle class on html and body elements
function toggleDocClass(thisClass: string) {
  const hasClass = document.documentElement.classList.contains(thisClass);
  if (hasClass) {
    removeDocClass(thisClass);
  } else {
    addDocClass(thisClass);
  }
}

// Function to toggle the theme
async function toggleTheme() {
  const newTheme = currentTheme.value === 'light' ? 'dark' : 'light';
  await SetTheme(newTheme); // Call Go backend to set the new theme
  currentTheme.value = newTheme;
}

// Watch for changes in the theme and update the body class
watch(currentTheme, (newTheme, oldTheme) => {
  if (oldTheme) {
    document.body.classList.remove(oldTheme);
  }
  document.body.classList.add(newTheme);
  // Also update the <html> element if needed
  document.documentElement.className = newTheme;

  // Re-initialize Mermaid with the correct theme
  mermaid.initialize({
    startOnLoad: false,
    theme: newTheme === 'dark' ? 'dark' : 'default',
  });
}, { immediate: true }); // immediate: true runs the watcher on component mount

// Print function
function printContent() {
    window.print();
}

// Save as PDF function (using browser's print to PDF)
function saveAsPDF(filePath: string) {
    // For web browsers, we can trigger print dialog with PDF option
    // The actual file saving needs to be handled differently in Wails
    window.print();
}

// Function to hide the modal
function hideHelpModal() {
    showHelpModal.value = false;
}

// Function to show the modal
function showHelpModalDialog(helpTitle: string, helpText: string) {
    if (helpRef.value) {
        helpRef.value.showHelpModalDialog(helpTitle, helpText);
        showHelpModal.value = true;
    }
}

// Settings dialog functions
function hideSettingsDialog() {
    showSettingsDialog.value = false;
}

function openSettingsDialog() {
    showSettingsDialog.value = true;
}

function onSettingsSaved() {
    // Settings have been saved successfully
    // You might want to emit an event or update the UI here
    console.log('Settings saved successfully');
}

onMounted(async () => {
  // Get initial theme from Go backend
  currentTheme.value = (await GetTheme()) as 'light' | 'dark';

  // Listen for theme changes initiated from the backend
  EventsOn('theme:changed', (newTheme: string) => {
    if (newTheme && (newTheme === 'light' || newTheme === 'dark')) {
      currentTheme.value = newTheme;
    }
  });

  // Listen for print events from Go backend
  EventsOn('print-content', () => {
      printContent();
  });

  EventsOn('save-as-pdf', (filePath: string) => {
      saveAsPDF(filePath);
  });

  // Listen for techdoc class manipulation events from Go backend
  EventsOn('add-doc-class', (thisClass: string) => {
    addDocClass(thisClass);
  });

  EventsOn('remove-doc-class', (thisClass: string) => {
    removeDocClass(thisClass);
  });

  EventsOn('toggle-doc-class', (thisClass: string) => {
    toggleDocClass(thisClass);
  });

  // Initialize Mermaid.js
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
  });

  // Listen for the 'markdown-rendered' event from the Go backend
  EventsOn('markdown-rendered', (html: string, title: string, date: string) => {
    console.log('Received markdownLoaded event. Updating HTML content.');
    renderedHTML.value = html;
    docHTMLTitle.value = title;
    docHTMLDate.value = date;
    errorMessage.value = ''; // Clear any previous error message
    nextTick(() => {
      console.log('Next tick after setting renderedHTML');
      // After the content is set, initialize Mermaid diagrams
      mermaid.run({
        nodes: document.querySelectorAll('.markdown-body .mermaid'),
      });
      document.title = title; // Set the document title
    });
  });
  EventsOn('error', (message: string) => {
    console.error('Received error event:', message);
    errorMessage.value = message; // Update the reactive error message.
    // Optionally, display the error directly within the main content area for visibility.
    renderedHTML.value = `<div style="color: #dc3545; padding: 20px; font-weight: bold; text-align: center;"><h1>An error occurred:</h1><p>${message}</p><p>Please try opening another file or check the file path.</p></div>`;
  });

  // Listen for an event from the Go backend to show the help dialog
  EventsOn("show-help", (helpTitle: string, helpText: string) => {
      console.log('Received show-help event:', helpTitle, helpText); // Debug log
      showHelpModalDialog(helpTitle, helpText);
  });

  // Listen for settings dialog events from Go backend
  EventsOn("show-settings", () => {
    console.log('Received show-settings event'); // Debug log
    openSettingsDialog();
  });
});

onUnmounted(() => {
  // Clean up the event listener when the component is destroyed
  EventsOff('markdown-rendered');
  EventsOff('error');
  EventsOff('add-doc-class');
  EventsOff('remove-doc-class');
  EventsOff('toggle-doc-class');
  EventsOff('print-content');
  EventsOff('save-as-pdf');
  EventsOff('show-help');
  EventsOff('show-settings');
});

</script>

<style>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.app-header {
  /* Fixed header below toolbar */
  flex-shrink: 0; /* Prevent header from shrinking */
  padding: 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  background-color: var(--header-bg, inherit);
  position: relative;
  z-index: 100;
}

.content-area {
  /* Scrollable content area */
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 20px;
  background-color: var(--content-bg, inherit);
}

/* Ensure markdown content doesn't have conflicting margins at top */
.content-area .markdown-body {
  margin-top: 0;
}

/* Theme-specific header styling */
.dark .app-header {
  border-bottom-color: rgba(255, 255, 255, 0.1);
  background-color: var(--header-bg-dark, inherit);
}

.light .app-header {
  border-bottom-color: rgba(0, 0, 0, 0.1);
  background-color: var(--header-bg-light, inherit);
}

/* Optional: Add subtle background differentiation */
.dark {
  --header-bg-dark: rgba(45, 55, 72, 0.8);
  --content-bg: inherit;
}

.light {
  --header-bg-light: rgba(248, 249, 250, 0.8);
  --content-bg: inherit;
}

/* Print-specific overrides to ensure proper printing */
@media print {
  .app-container {
    display: block !important;
    height: auto !important;
    overflow: visible !important;
  }

  .app-header {
    position: static !important;
    z-index: auto !important;
    flex-shrink: 0 !important;
  }

  .content-area {
    flex: none !important;
    overflow: visible !important;
    height: auto !important;
    max-height: none !important;
  }
}
</style>
