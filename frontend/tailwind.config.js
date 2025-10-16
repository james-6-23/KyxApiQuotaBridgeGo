/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主题色（浅色和深色模式通用）
        primary: {
          DEFAULT: '#1d9bf0',
          hover: '#1a8cd8',
          light: '#e3f2fd',
        },
        purple: {
          DEFAULT: '#7856ff',
          hover: '#6845e8',
          light: '#f3e8ff',
        },
        success: {
          DEFAULT: '#00ba7c',
          light: '#e8f5e9',
        },
        warning: {
          DEFAULT: '#f9a825',
          light: '#fff8e1',
        },
        error: {
          DEFAULT: '#f44336',
          light: '#ffebee',
        },
        // 深色主题颜色
        dark: {
          bg: '#000000',
          'bg-secondary': '#0a0a0a',
          'bg-tertiary': '#141414',
          'bg-hover': '#1a1a1a',
          border: '#262626',
          'border-hover': '#383838',
          text: '#e5e5e5',
          'text-secondary': '#a3a3a3',
          'text-tertiary': '#737373',
        },
        // 浅色主题颜色
        light: {
          bg: '#ffffff',
          'bg-secondary': '#f8f9fa',
          'bg-tertiary': '#f1f3f5',
          'bg-hover': '#e9ecef',
          border: '#dee2e6',
          'border-hover': '#ced4da',
          text: '#212529',
          'text-secondary': '#495057',
          'text-tertiary': '#6c757d',
        },
        // Grok 别名 (为了兼容现有代码)
        'grok-bg': '#000000',
        'grok-bg-secondary': '#0a0a0a',
        'grok-bg-tertiary': '#141414',
        'grok-bg-hover': '#1a1a1a',
        'grok-border': '#262626',
        'grok-border-hover': '#383838',
        'grok-text': '#e5e5e5',
        'grok-text-secondary': '#a3a3a3',
        'grok-text-tertiary': '#737373',
        'grok-primary': '#1d9bf0',
        'grok-purple': '#7856ff',
        'grok-success': '#00ba7c',
        'grok-warning': '#f9a825',
        'grok-error': '#f44336',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'sans-serif'],
        mono: ['SF Mono', 'Monaco', 'Cascadia Code', 'Roboto Mono', 'Courier New', 'monospace'],
      },
      boxShadow: {
        'grok': '0 0 20px rgba(29, 155, 240, 0.1)',
        'grok-lg': '0 0 40px rgba(29, 155, 240, 0.15)',
        'card': '0 1px 3px rgba(0, 0, 0, 0.5)',
        'card-hover': '0 4px 12px rgba(0, 0, 0, 0.6)',
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow': 'glow 2s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        glow: {
          '0%, 100%': { boxShadow: '0 0 20px rgba(29, 155, 240, 0.1)' },
          '50%': { boxShadow: '0 0 40px rgba(29, 155, 240, 0.3)' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [],
}
