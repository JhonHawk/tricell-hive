import { cn } from "cn";
import { Loader2Icon } from "lucide-react";
import { browserLanguage } from "../../language";
import { timelineCopy } from "../../timeline-copy";

function Spinner({ className, ...props }: React.ComponentProps<"svg">) {
  return (
    <Loader2Icon
      data-slot="spinner"
      role="status"
      aria-label={timelineCopy(browserLanguage()).loadingLabel}
      className={cn("size-4 animate-spin", className)}
      {...props}
    />
  );
}

export { Spinner };
