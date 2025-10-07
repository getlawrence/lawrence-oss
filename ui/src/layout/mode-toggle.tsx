import { Monitor, Moon, Sun } from "lucide-react";

import { useTheme } from "@/components/ThemeProvider";

interface ModeToggleProps {
  iconOnly?: boolean;
}

export function ModeToggle({ iconOnly = false }: ModeToggleProps) {
  const { theme, setTheme } = useTheme();

  const cycleTheme = () => {
    if (theme === "light") {
      setTheme("dark");
    } else if (theme === "dark") {
      setTheme("system");
    } else {
      setTheme("light");
    }
  };

  const getIcon = () => {
    switch (theme) {
      case "light":
        return <Sun className="h-4 w-4" />;
      case "dark":
        return <Moon className="h-4 w-4" />;
      case "system":
        return <Monitor className="h-4 w-4" />;
      default:
        return <Monitor className="h-4 w-4" />;
    }
  };

  return (
    <button
      onClick={cycleTheme}
      className="flex items-center gap-2 w-full px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground rounded-sm transition-colors"
    >
      {getIcon()}
      {!iconOnly && <span>Toggle theme</span>}
    </button>
  );
}
