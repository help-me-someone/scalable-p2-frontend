/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.tmpl"],
  theme: {
    extend: {
      colors: {
        'base': '#1e1e2e',
        'text-base': '#cdd6f4',
        'surface-2': '#585b70',
        'surface-1': '#45475a',
        'surface-0': '#313244',
      },
    },
  },
  plugins: [],
}

