<template>
  <div class="toggle-theme"><button @click="toggleTheme">Toggle Theme</button></div>
    <header class="app-header">
      <h1 v-html="docHTMLTitle" class="document-title"></h1>
      <p v-html="docHTMLDate" class="document-dates"></p>
      <!-- Display error message if preset -->
      <p v-if="errorMessage" class="error-message">Error: {{ errorMessage }}</p>
    </header>
    <main class="content-area">
      <article v-html="renderedHTML" id="content" class="markdown-body"></article>
    </main>
</template>

<script setup lang="ts">
import { ref, watch, nextTick,onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import { GetTheme, SetTheme } from '../wailsjs/go/main/App';
import mermaid from 'mermaid';

const renderedHTML = ref('<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>');
const docHTMLTitle = ref('');
const docHTMLDate = ref('');
const errorMessage = ref('');

// Get references to the modal elements
const helpModalOverlay = document.getElementById('help-modal-overlay');
const helpModalText = document.getElementById('help-modal-text');
const helpModalCloseBtn = document.getElementById('help-modal-close');

const currentTheme = ref('light');

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

onMounted(async () => {
  // Get initial theme from Go backend
  currentTheme.value = await GetTheme();

  // Listen for theme changes initiated from the backend
  EventsOn('theme:changed', (newTheme: string) => {
    if (newTheme) {
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
});

// Function to hide the modal
function hideHelpModal() {
    if (helpModalOverlay) {
        helpModalOverlay.style.display = 'none';
    }
}

// Function to show the modal
function showHelpModal(helpText: string) {
    if (helpModalOverlay && helpModalText) {
        helpModalText.textContent = helpText;
        helpModalOverlay.style.display = 'block';
    }
}

// Add event listeners to close the modal
if (helpModalCloseBtn) {
    helpModalCloseBtn.addEventListener('click', hideHelpModal);
}
// Also close if the user clicks on the dark overlay
if (helpModalOverlay) {
    helpModalOverlay.addEventListener('click', (event) => {
        if (event.target === helpModalOverlay) {
            hideHelpModal();
        }
    });
}

// Listen for an event from the Go backend to show the help dialog
EventsOn("show-help", (helpText) => {
    showHelpModal(helpText);
});

</script>

<style>
</style>
