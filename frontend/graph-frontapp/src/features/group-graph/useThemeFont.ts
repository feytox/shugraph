import { useMantineTheme } from '@mantine/core';
import { useQuery } from '@tanstack/react-query';

export function useThemeFont() {
  const { fontFamily } = useMantineTheme();
  return useQuery({
    queryKey: ['font', fontFamily],
    queryFn: () => document.fonts.load(`1em ${fontFamily}`),
    staleTime: Infinity,
    retry: false,
  });
}
