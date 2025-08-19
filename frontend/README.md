# Hyperlit Frontend

This is the Next.js + TypeScript + Tailwind CSS frontend for the Hyperlit copy trading platform.

## Stack

- [Next.js](https://nextjs.org/) (React framework)
- [TypeScript](https://www.typescriptlang.org/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Axios](https://axios-http.com/) (API calls)
- [Chart.js](https://www.chartjs.org/) via [react-chartjs-2](https://react-chartjs-2.js.org/)
- [Heroicons](https://heroicons.com/) (icons)

## Structure

- `pages/` - Page routes (index, dashboard, copy, vault, profile)
- `components/` - Reusable UI components
- `hooks/` - Custom React hooks
- `utils/` - Utility functions (API, etc.)
- `styles/` - Global styles

## Getting Started

```bash
cd frontend
npm install
npm run dev
```

## Customization

- Tailwind config in `tailwind.config.js`
- Global styles in `styles/globals.css`
- Layout in `components/Layout.tsx`