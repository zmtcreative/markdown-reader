import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import Toolbar from '../Toolbar.vue';

describe('Toolbar.vue', () => {
  it('emits toggle events from toolbar buttons', async () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'light',
        showFrontmatter: false,
      },
    });

    await wrapper.get('.refresh-btn').trigger('click');
    await wrapper.get('.frontmatter-toggle-btn').trigger('click');
    await wrapper.get('.theme-toggle-btn').trigger('click');

    expect(wrapper.emitted('refresh')).toHaveLength(1);
    expect(wrapper.emitted('toggleFrontmatter')).toHaveLength(1);
    expect(wrapper.emitted('toggleTheme')).toHaveLength(1);
  });

  it('updates the theme button title from the current theme', () => {
    const wrapper = mount(Toolbar, {
      props: {
        currentTheme: 'dark',
        showFrontmatter: true,
      },
    });

    expect(wrapper.get('.refresh-btn').attributes('title')).toBe('Refresh');
    expect(wrapper.get('.theme-toggle-btn').attributes('title')).toBe('Switch to light theme');
    expect(wrapper.get('.frontmatter-toggle-btn').attributes('title')).toBe('Hide Frontmatter');
  });
});