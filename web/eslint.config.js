import js from "@eslint/js";
import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import tseslint from "typescript-eslint";
import { defineConfig, globalIgnores } from "eslint/config";
import stylistic from "@stylistic/eslint-plugin";
import reactCompiler from "eslint-plugin-react-compiler";

export default defineConfig([
  globalIgnores(["dist"]),
  {
    files: ["**/*.{ts,tsx}"],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
      stylistic.configs.recommended,
    ],
    plugins: {
      "react-compiler": reactCompiler,
    },
    languageOptions: {
      globals: globals.browser,
    },
    rules: {
      // Add/override style rules here
      "@stylistic/comma-dangle": ["error", "always-multiline"],
      "@stylistic/object-curly-spacing": ["error", "always"],
      "@stylistic/array-bracket-spacing": ["error", "never"],
      "@stylistic/arrow-parens": ["error", "always"],
      "@stylistic/max-len": [
        "warn",
        {
          code: 120,
          ignoreUrls: true,
          ignoreStrings: true,
          ignoreTemplateLiterals: true,
        },
      ],
      "react-compiler/react-compiler": "error",
    },
  },
  {
    // The spice-panel component library legitimately exports non-components
    // (compound Object.assign roots, variant styles, the `toast` API, marker
    // slots, types), uses `_`-prefixed props kept for API compatibility, and
    // extend-only prop interfaces. Relax the dev-only fast-refresh + noise rules
    // here; correctness rules still apply.
    files: ["src/ui/**/*.{ts,tsx}", "src/dune-ui/**/*.{ts,tsx}", "src/gallery.tsx"],
    rules: {
      "react-refresh/only-export-components": "off",
      "@typescript-eslint/no-empty-object-type": "off",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_", destructuredArrayIgnorePattern: "^_" },
      ],
    },
  },
]);
