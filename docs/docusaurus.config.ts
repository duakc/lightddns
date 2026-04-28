import type { Config } from "@docusaurus/types";
import { themes as prismThemes } from "prism-react-renderer";
import { Options as DocsPluginOptions } from "@docusaurus/plugin-content-docs";
import { Options as PagePluginOptions } from "@docusaurus/plugin-content-pages";
import { Options as ThemeClassicOptions } from "@docusaurus/theme-classic";

const config: Config = {
  title: "Lightddns Documentation",
  tagline: "Lightddns Documentation",
  favicon: "img/favicon.ico",

  future: {
    v4: true,
  },

  url: "https://lightddns.duaky.com",
  baseUrl: "/",
  onBrokenLinks: "throw",
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },
  presets: [],
  plugins: [
    [
      "@docusaurus/plugin-content-docs",
      {
        path: "docs",
        sidebarPath: "./sidebars.ts",
        routeBasePath: "docs",
        include: ["**/*.md", "**/*.mdx"],
        sidebarCollapsed: true,
        disableVersioning: false,
        includeCurrentVersion: true,
      } satisfies DocsPluginOptions,
    ],
    [
      "@docusaurus/plugin-content-pages",
      {
        path: "src/pages",
      } satisfies PagePluginOptions,
    ],
    [
      "@docusaurus/theme-classic",
      {
        customCss: "./src/css/custom.css",
      } satisfies ThemeClassicOptions,
    ],
  ],
  themeConfig: {
    image: "img/docusaurus-social-card.jpg",
    colorMode: {
      respectPrefersColorScheme: true,
    },
    docs: {
      versionPersistence: "localStorage",
      sidebar: {
        hideable: true,
        autoCollapseCategories: false,
      },
    },
    navbar: {
      title: "Lightddns",
      items: [
        {
          label: "Documentation",
          type: "docSidebar",
          sidebarId: "documentationSidebar",
          position: "left",
        },
        {
          label: "Config Generator",
          to: "/configGenerator",
          position: "left",
        },
        {
          href: "https://github.com/duakc/lightddns",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  },
};

export default config;
