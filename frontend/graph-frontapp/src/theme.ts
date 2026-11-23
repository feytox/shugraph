import { createTheme, DEFAULT_THEME } from '@mantine/core';

const fontFamily = `'Golos Text Variable', ${DEFAULT_THEME.fontFamily}`;

export const theme = createTheme({ fontFamily, headings: { fontFamily } });
