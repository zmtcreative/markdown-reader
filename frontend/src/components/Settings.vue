<template>
  <Teleport to="body">
    <div
      id="settings-overlay"
      class="settings-overlay"
      v-show="show"
    >
      <div
        id="settings-dialog"
        class="settings-dialog"
        ref="settingsWindow"
      >
        <div class="settings-titlebar">
          <div
            class="window-title"
            @mousedown="startDrag"
          >
            <span>⚙️ Settings - Markdown Reader</span>
          </div>
          <div class="window-controls">
            <button
              id="settings-close"
              class="settings-close-button window-control-button"
              @click.stop="closeDialog"
              title="Close Settings"
            >
              &times;
            </button>
          </div>
        </div>
        <div class="settings-body">
          <h3 id="settings-title">Settings</h3>

          <!-- Tab Navigation -->
          <div class="tab-navigation">
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'markdown' }"
              @click="activeTab = 'markdown'"
            >
              📝 Markdown Processing
            </button>
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'alerts' }"
              @click="activeTab = 'alerts'"
            >
              ⚠️ Alert Callouts
            </button>
          </div>

          <form @submit.prevent="saveSettings" class="settings-form">
            <!-- Tab Content Area -->
            <div class="tab-content">

              <!-- Markdown Processing Tab -->
              <div v-show="activeTab === 'markdown'" class="tab-panel">
                <fieldset class="settings-section">
                  <legend>Markdown Processing</legend>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.allow_inline_html"
                      />
                      <span class="checkbox-custom"></span>
                      Allow Inline HTML
                    </label>
                    <p class="setting-description">
                      Allow raw HTML tags to be rendered in markdown content.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.sanitize_html"
                      />
                      <span class="checkbox-custom"></span>
                      Sanitize HTML
                    </label>
                    <p class="setting-description">
                      Remove potentially harmful HTML content and sanitize URLs for security.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.strip_h1"
                      />
                      <span class="checkbox-custom"></span>
                      Strip First H1 Heading
                    </label>
                    <p class="setting-description">
                      Remove the first H1 heading from markdown and use it as the document title.
                    </p>
                  </div>
                </fieldset>
              </div>

              <!-- Alert Callouts Tab -->
              <div v-show="activeTab === 'alerts'" class="tab-panel">
                <fieldset class="settings-section">
                  <legend>Alert Callouts</legend>

                  <div class="setting-group">
                    <label for="alert-callout-style" class="select-label">
                      Alert Callout Style:
                    </label>
                    <select
                      id="alert-callout-style"
                      v-model="localSettings.alert_callout_style"
                      class="setting-select"
                    >
                      <option
                        v-for="style in alertCalloutStyles"
                        :key="style"
                        :value="style"
                      >
                        {{ style }}
                      </option>
                    </select>
                    <p class="setting-description">
                      Choose the style for rendering alert callouts (notes, warnings, etc.).
                    </p>
                  </div>
                </fieldset>
              </div>

            </div>

            <!-- Buttons -->
            <div class="settings-buttons">
              <button type="button" @click="resetToDefaults" class="button button-secondary">
                Reset to Defaults
              </button>
              <button type="button" @click="closeDialog" class="button button-secondary">
                Cancel
              </button>
              <button type="submit" class="button button-primary">
                Save Settings
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { GetSettings, GetAlertCalloutStyles, SaveSettings } from '../../wailsjs/go/main/App';
import { app } from '../../wailsjs/go/models';

// Define the settings interface to match the generated TypeScript model
type Settings = app.Config;

// Component props
const props = defineProps<{
  show: boolean;
}>();

// Component emits
const emit = defineEmits<{
  close: [];
  saved: [];
}>();

// Reactive state
const localSettings = ref<Settings>(app.Config.createFrom({
  allow_inline_html: true,
  sanitize_html: true,
  strip_h1: true,
  alert_callout_style: 'GFMPlus'
}));

const alertCalloutStyles = ref<string[]>([]);
const loading = ref(false);

// Tab functionality
const activeTab = ref<'markdown' | 'alerts'>('markdown'); // Default to markdown tab

// Dialog window functionality
const settingsWindow = ref<HTMLElement>();
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// Load settings when dialog is shown
watch(() => props.show, async (newShow) => {
  if (newShow) {
    // Reset to the first tab when dialog opens
    activeTab.value = 'markdown';
    await loadSettings();
  }
});

// Load current settings and alert callout styles
async function loadSettings() {
  try {
    loading.value = true;

    // Load current settings
    const currentSettings = await GetSettings();
    localSettings.value = app.Config.createFrom(currentSettings);

    // Load available alert callout styles
    alertCalloutStyles.value = await GetAlertCalloutStyles();
  } catch (error) {
    console.error('Error loading settings:', error);
  } finally {
    loading.value = false;
  }
}

// Save settings
async function saveSettings() {
  try {
    loading.value = true;

    // Create proper Config object before saving
    const configToSave = app.Config.createFrom(localSettings.value);
    await SaveSettings(configToSave);

    emit('saved');
    closeDialog();
  } catch (error) {
    console.error('Error saving settings:', error);
    // You might want to show an error message to the user here
    alert('Error saving settings: ' + error);
  } finally {
    loading.value = false;
  }
}

// Reset to default values
function resetToDefaults() {
  localSettings.value = app.Config.createFrom({
    allow_inline_html: true,
    sanitize_html: true,
    strip_h1: true,
    alert_callout_style: 'GFMPlus'
  });
}

// Close dialog
function closeDialog() {
  emit('close');
}

function startDrag(event: MouseEvent) {
  isDragging.value = true;

  if (settingsWindow.value) {
    const rect = settingsWindow.value.getBoundingClientRect();
    dragOffset.value = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top
    };
  }

  document.addEventListener('mousemove', drag);
  document.addEventListener('mouseup', stopDrag);
  event.preventDefault();
  event.stopPropagation();
}

function drag(event: MouseEvent) {
  if (!isDragging.value || !settingsWindow.value) return;

  const x = event.clientX - dragOffset.value.x;
  const y = event.clientY - dragOffset.value.y;

  settingsWindow.value.style.left = `${x}px`;
  settingsWindow.value.style.top = `${y}px`;
}

function stopDrag() {
  isDragging.value = false;
  document.removeEventListener('mousemove', drag);
  document.removeEventListener('mouseup', stopDrag);
}

// Initialize when component mounts
loadSettings();
</script>

<style scoped>
/* Settings overlay - covers the entire screen transparently */
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 9999;
  background-color: transparent;
  pointer-events: all;
  font-size: 85%;
}

/* Settings dialog window */
.settings-dialog {
  position: fixed;
  top: 50px;
  left: 100px;
  width: 640px;
  max-height: 90vh;
  overflow: hidden;
  background-color: #2e3440;
  color: #d8dee9;
  border: 1px solid #4c566a;
  border-radius: 8px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  pointer-events: all;
  resize: both;
  min-width: 500px;
  min-height: 300px;
}

/* Settings titlebar */
.settings-titlebar {
  background-color: #2e3440;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #4c566a;
  user-select: none;
}

.window-title {
  font-weight: 600;
  color: #d8dee9;
  font-size: 0.9rem;
  cursor: move;
  flex: 1;
}

.window-controls {
  display: flex;
  gap: 4px;
}

.window-control-button {
  background: none;
  border: none;
  color: #d8dee9;
  font-size: 16px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 3px;
  transition: background-color 0.2s ease;
}

.window-control-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.settings-close-button.window-control-button:hover {
  background-color: #e74c3c;
  color: white;
}

/* Settings body */
.settings-body {
  padding: 0 20px 20px 20px;
  max-height: calc(90vh - 100px);
  overflow-y: auto;
}

/* Tab Navigation */
.tab-navigation {
  display: flex;
  margin: 1rem 0;
  border-bottom: 2px solid #4c566a;
  gap: 2px;
}

.tab-button {
  background-color: #3b4252;
  border: none;
  color: #d8dee9;
  padding: 0.75rem 1.25rem;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px 6px 0 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  position: relative;
  top: 2px; /* Align with border-bottom */
}

.tab-button:hover {
  background-color: #434c5e;
  color: #eceff4;
}

.tab-button.active {
  background-color: #5e81ac;
  color: white;
  border-bottom-color: #5e81ac;
}

/* Tab Content */
.tab-content {
  min-height: 200px;
}

.tab-panel {
  animation: fadeIn 0.2s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-section {
  border: 1px solid #4c566a;
  border-radius: 6px;
  padding: 1rem;
  margin: 0;
  background-color: #3b4252;
}

.settings-section legend {
  font-weight: bold;
  font-size: 1.1em;
  color: #d8dee9;
  padding: 0 0.5rem;
}

.setting-group {
  margin-bottom: 0.5rem;
}

.setting-group:last-child {
  margin-bottom: 0;
}

.checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  cursor: pointer;
  font-size: 0.9rem;
  color: #d8dee9;
  line-height: 1.5;
}

.checkbox-label input[type="checkbox"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.checkbox-custom {
  width: 12px;
  height: 12px;
  border: 2px solid #4c566a;
  border-radius: 3px;
  background-color: #2e3440;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 2px;
  transition: all 0.2s ease;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom {
  background-color: #5e81ac;
  border-color: #5e81ac;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom::after {
  content: '✓';
  color: white;
  font-size: 12px;
  font-weight: bold;
}

.checkbox-label:hover .checkbox-custom {
  border-color: #5e81ac;
}

.select-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #d8dee9;
}

.setting-select {
  width: 100%;
  max-width: 200px;
  padding: 0.5rem;
  border: 2px solid #4c566a;
  border-radius: 4px;
  background-color: #2e3440;
  color: #d8dee9;
  font-size: 0.85rem;
  margin-bottom: 0.25rem;
}

.setting-select:focus {
  outline: none;
  border-color: #5e81ac;
}

.setting-description {
  font-size: 0.875rem;
  color: #a3be8c;
  margin: 0.25rem 0 0 1.75rem;
  line-height: 1.1;
}

.settings-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #4c566a;
}

.button {
  padding: 0.7rem 1.25rem;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 600;
}

.button-primary {
  background-color: #5e81ac;
  color: white;
}

.button-primary:hover {
  background-color: #4c729a;
}

.button-secondary {
  background-color: #4c566a;
  color: #d8dee9;
}

.button-secondary:hover {
  background-color: #5d6b83;
}

.button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
