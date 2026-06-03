import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import FrontMatter from '../FrontMatter.vue';

describe('FrontMatter.vue', () => {
  it('renders frontmatter when visible and html content is present', () => {
    const wrapper = mount(FrontMatter, {
      props: {
        isVisible: true,
        frontmatterHTML: '<div class="frontmatter-container"><span class="fm-key">title</span></div>',
      },
    });

    expect(wrapper.find('.frontmatter-section').exists()).toBe(true);
    expect(wrapper.html()).toContain('fm-key');
  });

  it('hides the section when html content is empty', () => {
    const wrapper = mount(FrontMatter, {
      props: {
        isVisible: true,
        frontmatterHTML: '   ',
      },
    });

    expect(wrapper.find('.frontmatter-section').exists()).toBe(false);
  });
});