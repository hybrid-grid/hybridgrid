// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';

import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  vite: {
    plugins: [tailwindcss()],
  },
  integrations: [
    starlight({
      title: 'Hybrid-Grid',
      logo: {
        src: './src/assets/logo.svg',
        replacesTitle: false,
      },
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/h3nr1-d14z/hybridgrid' },
      ],
      customCss: ['./src/styles/starlight-theme.css'],
      editLink: {
        baseUrl: 'https://github.com/h3nr1-d14z/hybridgrid/edit/main/site/src/content/docs/',
      },
      sidebar: [
        {
          label: 'Start here',
          items: [
            { label: 'What is Hybrid-Grid', slug: 'docs' },
            { label: 'Quick start', slug: 'docs/quick-start' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'CLI reference', slug: 'docs/cli-reference' },
            { label: 'Configuration', slug: 'docs/configuration' },
            { label: 'Dashboard', slug: 'docs/dashboard' },
            { label: 'Flutter & Unity builds', slug: 'docs/flutter-and-unity' },
            { label: 'Docker & clustering', slug: 'docs/docker-and-clustering' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Architecture', slug: 'docs/architecture' },
            { label: 'Troubleshooting', slug: 'docs/troubleshooting' },
          ],
        },
      ],
    }),
  ],
});
