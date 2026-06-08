import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import React from 'react';
import { afterEach, vi } from 'vitest';

afterEach(() => {
  cleanup();
});

Element.prototype.scrollIntoView = vi.fn();

if (!HTMLFormElement.prototype.requestSubmit) {
  HTMLFormElement.prototype.requestSubmit = function requestSubmit() {
    this.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
  };
}

vi.mock('framer-motion', () => ({
  motion: new Proxy(
    {},
    {
      get: (_target, tag) => {
        if (typeof tag !== 'string') {
          return undefined;
        }

        const MotionComponent = React.forwardRef<HTMLElement, Record<string, unknown> & { children?: React.ReactNode }>(
          ({ children, ...props }, ref) => {
            const {
              animate,
              exit,
              initial,
              transition,
              variants,
              viewport,
              whileHover,
              whileInView,
              whileTap,
              ...rest
            } = props;

            return React.createElement(tag, { ...rest, ref }, children);
          },
        );
        MotionComponent.displayName = `motion.${tag}`;
        return MotionComponent;
      },
    },
  ),
}));
