/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./frontend/templates/**/*.{html,templ}",
    "./internal/**/*.{go,templ}",
    "./**/*.{go,templ}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
