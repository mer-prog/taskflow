// Root layout is handled by [locale]/layout.tsx via next-intl middleware
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
