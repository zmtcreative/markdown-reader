<template>
  <!-- Help Modal HTML Start -->
  <Teleport to="body">
  <div id="help-modal-overlay" class="help-overlay" v-show="show" @click="onModalOverlayClick">
      <div id="help-modal-content" class="help-dialog">
          <div class="help-titlebar">
              <div class="window-title">
                  <span>{{ helpModalTitle || 'Help' }}</span>
              </div>
              <div class="window-controls">
                  <button id="help-modal-close" class="help-close-button window-control-button" @click="hideHelpModal" title="Close Help">&times;</button>
              </div>
          </div>
          <div class="help-body">
              <div id="help-modal-text" v-html="helpModalText"></div>
          </div>
      </div>
  </div>
  </Teleport>
  <!-- Help Modal HTML End -->
</template>

<script setup lang="ts">
import { ref } from 'vue';

// Props
interface Props {
  show: boolean;
}

const props = defineProps<Props>();

// Emits
const emit = defineEmits<{
  close: [];
}>();

// Modal reactive variables
const helpModalTitle = ref('');
const helpModalText = ref('');

// Function to hide the modal
function hideHelpModal() {
    emit('close');
}

// Function to show the modal with content
function showHelpModalDialog(helpTitle: string, helpText: string) {
    helpModalTitle.value = helpTitle;
    helpModalText.value = helpText;
}

// Handle clicking on modal overlay (close modal)
function onModalOverlayClick(event: Event) {
    if (event.target === event.currentTarget) {
        hideHelpModal();
    }
}

// Expose the showHelpModalDialog function so parent can call it
defineExpose({
  showHelpModalDialog
});
</script>

<style scoped>
/* CSS Variables - Define all the custom properties used in this component */
/* .help-overlay {
  --dialog-color-bg-primary: #ffffff;
  --dialog-color-bg-secondary: #f8fafc;
  --dialog-color-bg-tertiary: #f1f5f9;
  --dialog-color-bg-hover: #f8fafc;
  --dialog-color-border: #e2e8f0;
  --dialog-color-border-light: #f1f5f9;
  --dialog-color-text-primary: #1e293b;
  --dialog-color-text-secondary: #64748b;
  --dialog-color-accent: #3b82f6;
  --dialog-color-accent-hover: #2563eb;
  --dialog-font-family-sans: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
} */

/* Help overlay - covers the entire screen transparently */
.help-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(3px);
}

/* Help dialog window */
.help-dialog {
  background: var(--dialog-color-bg-primary);
  border: 1px solid var(--dialog-color-border);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  min-width: 500px;
  max-width: 800px;
  width: 90vw;
  max-height: 85vh;
  position: relative;
  display: flex;
  flex-direction: column;
  font-family: var(--dialog-font-family-sans);
}

/* Titlebar */
.help-titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--dialog-color-bg-secondary);
  border-bottom: 1px solid var(--dialog-color-border);
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  user-select: none;
}

.window-title {
  color: var(--dialog-color-text-primary);
  font-weight: 600;
  font-size: 14px;
  flex: 1;
}

.window-controls {
  display: flex;
  gap: 8px;
}

.window-control-button {
  background: none;
  border: none;
  color: var(--dialog-color-text-secondary);
  font-size: 18px;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.window-control-button:hover {
  background: var(--dialog-color-bg-tertiary);
  color: var(--dialog-color-text-primary);
}

.help-close-button:hover {
  background: #e74c3c;
  color: white;
}

/* Help body */
.help-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
  color: var(--dialog-color-text-primary);
  line-height: 1.6;
}

/* Content styling */
.help-body h1,
.help-body h2,
.help-body h3,
.help-body h4,
.help-body h5,
.help-body h6 {
  color: var(--dialog-color-text-primary);
  margin: 1.2em 0 0.8em 0;
  font-weight: 600;
}

.help-body h1:first-child,
.help-body h2:first-child,
.help-body h3:first-child {
  margin-top: 0;
}

.help-body h1 { font-size: 1.8em; }
.help-body h2 { font-size: 1.5em; }
.help-body h3 { font-size: 1.3em; }
.help-body h4 { font-size: 1.1em; }

.help-body p {
  margin: 0 0 1em 0;
  color: var(--dialog-color-text-primary);
}

.help-body a {
  color: var(--dialog-color-accent);
  text-decoration: none;
  transition: color 0.2s ease;
}

.help-body a:hover {
  color: var(--dialog-color-accent-hover);
  text-decoration: underline;
}

.dark .help-body table.dialog-data tr td p:has(a) {
    background: white;
}

.help-body ul,
.help-body ol {
  margin: 0 0 1em 0;
  padding-left: 1.5em;
}

.help-body li {
  margin-bottom: 0.5em;
}

.help-body code {
  background: var(--dialog-color-bg-secondary);
  color: var(--dialog-color-text-primary);
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 0.9em;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.help-body pre {
  background: var(--dialog-color-bg-secondary);
  color: var(--dialog-color-text-primary);
  padding: 16px;
  border-radius: 6px;
  border: 1px solid var(--dialog-color-border);
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9em;
  line-height: 1.5;
}

.help-body pre code {
  background: none;
  padding: 0;
}

.help-body blockquote {
  background: var(--dialog-color-bg-secondary);
  border-left: 4px solid var(--dialog-color-accent);
  margin: 1em 0;
  padding: 12px 16px;
  border-radius: 0 6px 6px 0;
}

.help-body hr {
  border: none;
  border-top: 1px solid var(--dialog-color-border);
  margin: 2em 0;
}

.help-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
}

.help-body th,
.help-body td {
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid var(--dialog-color-border);
}

.help-body th {
  background: var(--dialog-color-bg-secondary);
  font-weight: 600;
}

/* Scrollbar styling */
.help-body::-webkit-scrollbar {
  width: 8px;
}

.help-body::-webkit-scrollbar-track {
  background: var(--dialog-color-bg-secondary);
  border-radius: 4px;
}

.help-body::-webkit-scrollbar-thumb {
  background: var(--dialog-color-border);
  border-radius: 4px;
}

.help-body::-webkit-scrollbar-thumb:hover {
  background: var(--dialog-color-text-secondary);
}

/* Responsive adjustments */
@media (max-width: 600px) {
  .help-dialog {
    min-width: unset;
    width: 95vw;
    margin: 20px;
  }

  .help-body {
    padding: 16px;
  }
}

/* Animation */
.help-overlay {
  animation: fadeIn 0.2s ease-in-out;
}

.help-dialog {
  animation: slideIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(-20px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
</style>
