const colors = require('tailwindcss/colors');
const layerstack = require('@layerstack/tailwind/plugin');

module.exports = {
  content: [
    './src/**/*.{html,svelte}', 
    './node_modules/svelte-ux/**/*.{svelte,js}'
  ],

  // See customization docs: https://svelte-ux.techniq.dev/customization
  ux: {
    themes: {
      "light": {
        "color-scheme": "light",
        "primary": "hsl(257.4075 100% 50%)",
        "secondary": "hsl(310.4453 100% 50%)",
        "accent": "hsl(173.4835 100% 42.1865%)",
        "neutral": "hsl(214.2857 19.6262% 20.9804%)",
        "surface-100": "hsl(180 100% 100%)",
        "surface-200": "hsl(0 0% 94.902%)",
        "surface-300": "hsl(180 1.9608% 90%)"
      },
      "dark": {
        "color-scheme": "dark",
        "primary": "hsl(234.8208 100% 72.6713%)",
        "secondary": "hsl(313.3209 100% 66.1653%)",
        "accent": "hsl(173.7346 100% 40.1728%)",
        "neutral": "hsl(213.3333 17.6471% 20%)",
        "surface-100": "hsl(212.3077 18.3099% 13.9216%)",
        "surface-200": "hsl(212.7273 18.0328% 11.9608%)",
        "surface-300": "hsl(213.3333 17.6471% 10%)"
      }
    }
  },

  plugins: [
    layerstack,  // uses hsl() color space by default. To use oklch(), use: layerstack({ colorSpace: 'oklch' }),
  ]
};