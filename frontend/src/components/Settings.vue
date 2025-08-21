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
              :class="{ 'active': activeTab === 'application' }"
              @click="activeTab = 'application'"
            >
              🏠 Application
            </button>
            <button
              type="button"
              class="tab-button"
              :class="{ 'active': activeTab === 'markdown' }"
              @click="activeTab = 'markdown'"
            >
              📝 Markdown Options
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

              <!-- Application Tab -->
              <div v-show="activeTab === 'application'" class="tab-panel">
                <fieldset class="settings-section">
                  <legend>Application Settings</legend>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_inline_html"
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
                        v-model="localSettings.application.use_sanitize_html"
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
                        v-model="localSettings.application.strip_h1"
                      />
                      <span class="checkbox-custom"></span>
                      Strip First H1 Heading
                    </label>
                    <p class="setting-description">
                      Remove the first H1 heading from markdown and use it as the document title.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.application.use_frontmatter"
                      />
                      <span class="checkbox-custom"></span>
                      Parse Frontmatter
                    </label>
                    <p class="setting-description">
                      Extract and parse YAML frontmatter from markdown files.
                    </p>
                  </div>
                </fieldset>
              </div>

              <!-- Markdown Options Tab -->
              <div v-show="activeTab === 'markdown'" class="tab-panel">
                <fieldset class="settings-section">
                  <legend>Markdown Processing Options</legend>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_gfm"
                      />
                      <span class="checkbox-custom"></span>
                      GitHub Flavored Markdown
                    </label>
                    <p class="setting-description">
                      Enable GitHub Flavored Markdown and PHP Markdown Extensions.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_emoji"
                      />
                      <span class="checkbox-custom"></span>
                      Emoji Support
                    </label>
                    <p class="setting-description">
                      Convert emoji codes (like :smile:) to emoji characters.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_mermaid"
                      />
                      <span class="checkbox-custom"></span>
                      Mermaid Diagrams
                    </label>
                    <p class="setting-description">
                      Enable Mermaid diagram rendering for charts and graphs.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_figure"
                      />
                      <span class="checkbox-custom"></span>
                      Image Figure Wrapping
                    </label>
                    <p class="setting-description">
                      Wrap images in figure elements with captions.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_anchor"
                      />
                      <span class="checkbox-custom"></span>
                      Anchor Links on Headings
                    </label>
                    <p class="setting-description">
                      Add clickable anchor links to heading elements.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_fences"
                      />
                      <span class="checkbox-custom"></span>
                      Fenced DIVs
                    </label>
                    <p class="setting-description">
                      Enable fenced div blocks for custom content containers.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_sections"
                      />
                      <span class="checkbox-custom"></span>
                      Wrap Headings in SECTION Elements
                    </label>
                    <p class="setting-description">
                      Wrap headings and their content in HTML section elements.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_highlighting"
                      />
                      <span class="checkbox-custom"></span>
                      Fenced Code Highlighting
                    </label>
                    <p class="setting-description">
                      Enable syntax highlighting for code blocks.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_fancylists"
                      />
                      <span class="checkbox-custom"></span>
                      Pandoc-Style Fancy Lists
                    </label>
                    <p class="setting-description">
                      Allow advanced list formatting with custom numbering.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_attributes"
                      />
                      <span class="checkbox-custom"></span>
                      Custom Attributes
                    </label>
                    <p class="setting-description">
                      Allow custom attributes using '{.myclass}' syntax.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.markdown.use_typographic"
                      />
                      <span class="checkbox-custom"></span>
                      Typographic Extensions
                    </label>
                    <p class="setting-description">
                      Use fancy quotes and typographic enhancements.
                    </p>
                  </div>
                </fieldset>
              </div>

              <!-- Alert Callouts Tab -->
              <div v-show="activeTab === 'alerts'" class="tab-panel">
                <fieldset class="settings-section">
                  <legend>Alert Callouts</legend>

                  <div class="setting-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        v-model="localSettings.alert_callouts.use_alertcallouts"
                      />
                      <span class="checkbox-custom"></span>
                      Enable Alert Callouts
                    </label>
                    <p class="setting-description">
                      Enable GitHub and/or Obsidian style alert/callout blocks.
                    </p>
                  </div>

                  <div class="setting-group">
                    <label for="alert-callout-style" class="select-label">
                      Alert Callout Style:
                    </label>
                    <select
                      id="alert-callout-style"
                      v-model="localSettings.alert_callouts.alertcallout_style"
                      class="setting-select"
                    >
                      <option
                        v-for="(description, key) in alertCalloutStyles"
                        :key="key"
                        :value="key"
                      >
                        {{ description }}
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
              <button type="button" @click="saveSettingsSessionOnly" class="button button-info">
                Apply for Session
              </button>
              <button type="submit" class="button button-primary">
                Save to Disk
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
import { GetSettings, GetAlertCalloutStyles, SaveSettings, SaveSettingsSessionOnly } from '../../wailsjs/go/main/App';
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

// Create default settings structure matching the new nested format
const createDefaultSettings = (): Settings => {
  return app.Config.createFrom({
    application: {
      use_inline_html: true,
      use_sanitize_html: true,
      strip_h1: true,
      use_frontmatter: true
    },
    markdown: {
      use_gfm: true,
      use_emoji: true,
      use_mermaid: true,
      use_figure: true,
      use_anchor: true,
      use_fences: true,
      use_sections: true,
      use_highlighting: true,
      use_fancylists: true,
      use_attributes: true,
      use_typographic: true
    },
    alert_callouts: {
      use_alertcallouts: true,
      alertcallout_style: 'GFMPlus'
    }
  });
};

// Reactive state
const localSettings = ref<Settings>(createDefaultSettings());
const alertCalloutStyles = ref<Record<string, string>>({});
const loading = ref(false);

// Tab functionality - Default to application tab
const activeTab = ref<'application' | 'markdown' | 'alerts'>('application');

// Dialog window functionality
const settingsWindow = ref<HTMLElement>();
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// Load settings when dialog is shown
watch(() => props.show, async (newShow) => {
  if (newShow) {
    // Reset to the first tab when dialog opens
    activeTab.value = 'application';
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

// Save settings for session only (don't write to disk)
async function saveSettingsSessionOnly() {
  try {
    loading.value = true;

    // Create proper Config object before saving
    const configToSave = app.Config.createFrom(localSettings.value);
    await SaveSettingsSessionOnly(configToSave);

    emit('saved');
    closeDialog();
  } catch (error) {
    console.error('Error applying settings for session:', error);
    // You might want to show an error message to the user here
    alert('Error applying settings for session: ' + error);
  } finally {
    loading.value = false;
  }
}

// Reset to default values
function resetToDefaults() {
  localSettings.value = createDefaultSettings();
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
/* CSS Variables - Define all the custom properties used in this component */
/* .settings-overlay {
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
  --font-family-sans: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
} */

/* Settings overlay - covers the entire screen transparently */
.settings-overlay {
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

/* Settings dialog window */
.settings-dialog {
  background: var(--dialog-color-bg-primary);
  border: 1px solid var(--dialog-color-border);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  min-width: 500px;
  max-width: 700px;
  width: 90vw;
  max-height: 85vh;
  position: relative;
  display: flex;
  flex-direction: column;
  font-family: var(--dialog-font-family-sans);
}

/* Titlebar */
.settings-titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--dialog-color-bg-secondary);
  border-bottom: 1px solid var(--dialog-color-border);
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  cursor: move;
  user-select: none;
}

.window-title {
  color: var(--dialog-color-text-primary);
  font-weight: 600;
  font-size: 14px;
  flex: 1;
  cursor: move;
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

.settings-close-button:hover {
  background: #e74c3c;
  color: white;
}

/* Settings body */
.settings-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

#settings-title {
  margin: 0 0 20px 0;
  color: var(--dialog-color-text-primary);
  font-size: 18px;
  font-weight: 600;
}

/* Tab navigation */
.tab-navigation {
  display: flex;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--dialog-color-border);
  gap: 1px;
}

.tab-button {
  background: var(--dialog-color-bg-secondary);
  border: none;
  color: var(--dialog-color-text-secondary);
  padding: 12px 16px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  border-radius: 6px 6px 0 0;
  position: relative;
  bottom: -1px;
}

.tab-button:hover {
  background: var(--dialog-color-bg-tertiary);
  color: var(--dialog-color-text-primary);
}

.tab-button.active {
  background: var(--dialog-color-bg-primary);
  color: var(--dialog-color-tab-accent);
  border: 2px solid var(--dialog-color-bg-secondary);
  border-bottom: 2px solid var(--dialog-color-accent);

}

/* Tab content */
.tab-content {
  min-height: 300px;
}

.tab-panel {
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Settings form */
.settings-form {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.settings-section {
  border: 1px solid var(--dialog-color-border);
  border-radius: 6px;
  margin-bottom: 20px;
  padding: 0;
  background: var(--dialog-color-bg-primary);
}

.settings-section legend {
  color: var(--dialog-color-text-primary);
  font-weight: 600;
  font-size: 16px;
  padding: 0 12px;
  margin-left: 8px;
}

.settings-section > div:first-of-type {
  padding-top: 16px;
}

.settings-section > div:last-of-type {
  padding-bottom: 16px;
}

/* Setting groups */
.setting-group {
  padding: 12px 20px;
  border-bottom: 1px solid var(--dialog-color-border-light);
}

.setting-group:last-child {
  border-bottom: none;
}

.setting-group:hover {
  background: var(--dialog-color-bg-hover);
}

/* Checkbox styling */
.checkbox-label {
  display: flex;
  align-items: center;
  color: var(--dialog-color-text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  gap: 12px;
  line-height: 1.4;
}

.checkbox-label input[type="checkbox"] {
  display: none;
}

.checkbox-custom {
  width: 18px;
  height: 18px;
  border: 2px solid var(--dialog-color-border);
  border-radius: 3px;
  background: var(--dialog-color-bg-primary);
  position: relative;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom {
  background: var(--dialog-color-accent);
  border-color: var(--dialog-color-accent);
}

.checkbox-custom::after {
  content: '';
  position: absolute;
  width: 5px;
  height: 9px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  top: 1px;
  left: 5px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom::after {
  opacity: 1;
}

.checkbox-custom:hover {
  border-color: var(--dialog-color-tab-accent);
}

/* Select styling */
.select-label {
  display: block;
  color: var(--dialog-color-text-primary);
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
}

.setting-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--dialog-color-border);
  border-radius: 4px;
  background: var(--dialog-color-bg-primary);
  color: var(--dialog-color-text-primary);
  font-size: 14px;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.setting-select:hover {
  border-color: var(--dialog-color-accent);
}

.setting-select:focus {
  outline: none;
  border-color: var(--dialog-color-accent);
  box-shadow: 0 0 0 2px rgba(74, 144, 226, 0.2);
}

/* Setting descriptions */
.setting-description {
  margin: 8px 0 0 30px;
  color: var(--dialog-color-text-secondary);
  font-size: 12px;
  line-height: 1.4;
}

/* Buttons */
.settings-buttons {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding-top: 20px;
  border-top: 1px solid var(--dialog-color-border);
  margin-top: auto;
}

.button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 80px;
}

.button-primary {
  background: var(--dialog-color-accent);
  color: white;
}

.button-primary:hover {
  background: var(--dialog-color-accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.3);
}

.button-secondary {
  background: var(--dialog-color-bg-secondary);
  color: var(--dialog-color-text-primary);
  border: 1px solid var(--dialog-color-border);
}

.button-secondary:hover {
  background: var(--dialog-color-bg-tertiary);
  transform: translateY(-1px);
}

.button-info {
  background: #10b981;
  color: white;
  border: 1px solid #059669;
}

.button-info:hover {
  background: #059669;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.button:active {
  transform: translateY(0);
}

/* Scrollbar styling */
.settings-body::-webkit-scrollbar {
  width: 8px;
}

.settings-body::-webkit-scrollbar-track {
  background: var(--dialog-color-bg-secondary);
  border-radius: 4px;
}

.settings-body::-webkit-scrollbar-thumb {
  background: var(--dialog-color-border);
  border-radius: 4px;
}

.settings-body::-webkit-scrollbar-thumb:hover {
  background: var(--dialog-color-text-secondary);
}

/* Responsive adjustments */
@media (max-width: 600px) {
  .settings-dialog {
    min-width: unset;
    width: 95vw;
    margin: 20px;
  }

  .tab-navigation {
    flex-direction: column;
  }

  .tab-button {
    border-radius: 4px;
    margin-bottom: 2px;
  }

  .settings-buttons {
    flex-direction: column-reverse;
  }

  .button {
    width: 100%;
  }
}
</style>
