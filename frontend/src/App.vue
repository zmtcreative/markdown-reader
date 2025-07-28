<template>
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
import { ref, nextTick,onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import mermaid from 'mermaid';

const renderedHTML = ref('<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>');
const docHTMLTitle = ref('');
const docHTMLDate = ref('');

const errorMessage = ref('');

// Get references to the modal elements
const helpModalOverlay = document.getElementById('help-modal-overlay');
const helpModalText = document.getElementById('help-modal-text');
const helpModalCloseBtn = document.getElementById('help-modal-close');

onMounted(() => {
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
