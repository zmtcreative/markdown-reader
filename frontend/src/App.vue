<template>
  <article v-html="renderedHTML" class="markdown-body"></article>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

const renderedHTML = ref('<h3>No markdown file specified. Please open a markdown file using File > Open.</h3>');

onMounted(async () => {
  // Get the markdown file path from the command line arguments
  const args = await window.go.main.App.GetArgs();
  console.log('Command line arguments length:', args.length);
  console.log('Command line arguments:', args);
  if (args.length > 0) {
    const filePath = args[0];
    try {
      renderedHTML.value = await window.go.main.App.ProcessMarkdown(filePath);
    } catch (e) {
      console.error(e);
      renderedHTML.value = 'Error rendering markdown.';
    }
  } else {
    renderedHTML.value = 'No markdown file specified.';
  }
});

// Function to update the rendered HTML
const updateContent = (html: string) => {
  renderedHTML.value = html;
};

onMounted(() => {
  // Listen for the 'markdown-rendered' event from the Go backend
  EventsOn('markdown-rendered', updateContent);
});

onUnmounted(() => {
  // Clean up the event listener when the component is destroyed
  EventsOff('markdown-rendered');
});

</script>

<style>
</style>
