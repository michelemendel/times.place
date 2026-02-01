/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    container: {
      center: true,
      padding: {
        DEFAULT: '1rem',
        sm: '1rem',
        md: '1rem',
        lg: '1rem',
        xl: '1rem',
        '2xl': '1rem',
      },
      screens: {
        // Don't apply fixed width below md so mobile uses full width
        // sm: '640px', // removed
        md: '768px',
        lg: '1024px',
        xl: '1280px',
        '2xl': '1280px', // Cap max width for readability
      },
    },
    extend: {}
  },
  plugins: []
};

