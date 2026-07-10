import { useCallback, useEffect, useRef, useState } from "react";

export function useCopyCommand(command: string): {
  copied: boolean;
  copy: () => Promise<void>;
} {
  const [copied, setCopied] = useState(false);
  const resetTimer = useRef<number | null>(null);

  const copy = useCallback(async () => {
    await navigator.clipboard.writeText(command);
    setCopied(true);

    if (resetTimer.current !== null) {
      window.clearTimeout(resetTimer.current);
    }

    resetTimer.current = window.setTimeout(() => setCopied(false), 2_000);
  }, [command]);

  useEffect(
    () => () => {
      if (resetTimer.current !== null) {
        window.clearTimeout(resetTimer.current);
      }
    },
    [],
  );

  return { copied, copy };
}
