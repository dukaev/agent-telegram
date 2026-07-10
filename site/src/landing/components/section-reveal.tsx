import { motion } from "motion/react";

import { useReducedMotionPreference } from "../use-reduced-motion";

type SectionRevealProps = {
  children: React.ReactNode;
  className?: string;
  delay?: number;
};

export function SectionReveal({
  children,
  className = "",
  delay = 0,
}: SectionRevealProps): React.JSX.Element {
  const reduceMotion = useReducedMotionPreference();

  return (
    <motion.div
      className={className}
      initial={reduceMotion ? false : { opacity: 0, transform: "translateY(22px)" }}
      transition={{ delay, duration: 0.58, ease: [0.23, 1, 0.32, 1] }}
      viewport={{ amount: 0.2, once: true }}
      whileInView={{ opacity: 1, transform: "translateY(0)" }}
    >
      {children}
    </motion.div>
  );
}
