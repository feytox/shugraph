import { useMantineTheme } from '@mantine/core';
import { useQuery } from '@tanstack/react-query';

const latinCyrillicAndRubleSign = 'Aа₽';

export function useThemeFont() {
  const { fontFamily } = useMantineTheme();
  return useQuery({
    queryKey: ['font', fontFamily],
    queryFn: () => document.fonts.load(`1em ${fontFamily}`, latinCyrillicAndRubleSign),
    staleTime: Infinity,
    retry: false,
    networkMode: 'always',
  });
}
