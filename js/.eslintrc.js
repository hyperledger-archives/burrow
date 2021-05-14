module.exports = {
  env: {
    es6: true,
    node: true,
    mocha: true,
  },
  extends: ['plugin:@typescript-eslint/recommended'],
  globals: {
    Atomics: 'readonly',
    SharedArrayBuffer: 'readonly',
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2018,
    sourceType: 'module',
  },
  plugins: ['@typescript-eslint', 'prettier'],
  rules: {
    'prettier/prettier': [
      'error',
      {
        printWidth: 120,
        singleQuote: true,
        useTabs: false,
        tabWidth: 2,
        trailingComma: 'all',
      },
    ],
    '@typescript-eslint/no-namespace': 'off',
    curly: 2,
  },
};
