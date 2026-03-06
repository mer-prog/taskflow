import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { Button } from "@/components/ui/button";

export default function LandingPage() {
  const t = useTranslations("landing");
  const tc = useTranslations("common");

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          <span className="font-bold text-lg">{tc("appName")}</span>
          <Link href="/login">
            <Button variant="ghost" size="sm" className="min-h-[44px]">
              {t("login")}
            </Button>
          </Link>
        </div>
      </header>
      <main className="flex flex-1 items-center justify-center px-4">
        <div className="mx-auto max-w-2xl text-center">
          <h1 className="text-3xl font-bold tracking-tight sm:text-5xl">
            {t("heroTitle")}
          </h1>
          <p className="mt-6 text-lg text-muted-foreground">
            {t("heroDescription")}
          </p>
          <div className="mt-10 flex flex-col gap-3 sm:flex-row sm:justify-center">
            <Link href="/register">
              <Button size="lg" className="w-full sm:w-auto min-h-[44px]">
                {t("getStarted")}
              </Button>
            </Link>
            <Link href="/login">
              <Button variant="outline" size="lg" className="w-full sm:w-auto min-h-[44px]">
                {t("login")}
              </Button>
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
