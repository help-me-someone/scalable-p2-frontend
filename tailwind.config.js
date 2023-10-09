/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.tmpl"],
  theme: {
    extend: {
      colors: {
        'base': '#1e1e2e',
        'text-base': '#cdd6f4',
      },
    },
  },
  plugins: [],
}

