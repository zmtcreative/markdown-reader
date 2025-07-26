<template>
  <article v-html="renderedHTML" class="markdown-body"></article>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const renderedHTML = ref('');

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
</script>

<style>
</style>
