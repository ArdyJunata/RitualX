/**
 * Lighthouse CI configuration
 */
const PORT = 3000;
const BASE_URL = `http://localhost:${PORT}`;
const MARKETING_URLS = [`${BASE_URL}/`];

const LCP_BUDGET_MS = 2500;
const INP_BUDGET_MS = 200;
const CLS_BUDGET = 0.1;

module.exports = {
  ci: {
    collect: {
      startServerCommand: `npm run start -- --port ${PORT}`,
      startServerReadyPattern: "Ready in",
      startServerReadyTimeout: 120000,
      url: MARKETING_URLS,
      numberOfRuns: 3,
      settings: {
        preset: process.env.LHCI_FORM_FACTOR === "desktop" ? "desktop" : undefined,
        onlyCategories: [
          "performance",
          "seo",
          "accessibility",
          "best-practices",
        ],
      },
    },
    assert: {
      aggregationMethod: "median-run",
      assertions: {
        "largest-contentful-paint": ["error", { maxNumericValue: LCP_BUDGET_MS }],
        "cumulative-layout-shift": ["error", { maxNumericValue: CLS_BUDGET }],
        "total-blocking-time": ["error", { maxNumericValue: INP_BUDGET_MS }],
        "interaction-to-next-paint": ["warn", { maxNumericValue: INP_BUDGET_MS }],
        "categories:performance": ["error", { minScore: 0.9 }],
        "categories:seo": ["error", { minScore: 0.95 }],
        "categories:accessibility": ["error", { minScore: 0.95 }],
        "categories:best-practices": ["error", { minScore: 0.9 }],
      },
    },
    upload: {
      target: "filesystem",
      outputDir: "./.lighthouseci",
    },
  },
};
