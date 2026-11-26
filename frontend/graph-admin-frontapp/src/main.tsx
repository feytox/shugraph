import React from 'react';
import ReactDOM from 'react-dom/client';
import { createTheme, DEFAULT_THEME, MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import '@mantine/core/styles.css';
import '@fontsource-variable/golos-text';
import { App } from './App';

const fontFamily = `'Golos Text Variable', ${DEFAULT_THEME.fontFamily}`;
const theme = createTheme({ fontFamily, headings: { fontFamily } });
const queryClient = new QueryClient();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <MantineProvider theme={theme}>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </MantineProvider>
  </React.StrictMode>
);
