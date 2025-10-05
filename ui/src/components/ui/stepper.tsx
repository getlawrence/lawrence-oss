import { Check } from 'lucide-react';

import { cn } from '@/lib/utils';

export interface Step {
  id: string;
  title: string;
  description?: string;
}

export interface StepperProps {
  steps: Step[];
  currentStep: number;
  className?: string;
}

export function Stepper({ steps, currentStep, className }: StepperProps) {
  return (
    <div className={cn('w-full', className)}>
      <div className="flex items-center justify-between">
        {steps.map((step, index) => {
          const stepNumber = index + 1;
          const isCompleted = stepNumber < currentStep;
          const isCurrent = stepNumber === currentStep;
          const isUpcoming = stepNumber > currentStep;

          return (
            <div key={step.id} className="flex items-center flex-1">
              <div className="flex flex-col items-center">
                <div
                  className={cn(
                    'flex items-center justify-center w-6 h-6 rounded-full border-2 text-xs font-medium transition-colors',
                    {
                      'bg-primary border-primary text-primary-foreground': isCompleted || isCurrent,
                      'bg-background border-muted-foreground text-muted-foreground': isUpcoming,
                    }
                  )}
                >
                  {isCompleted ? <Check className="w-3 h-3" /> : <span>{stepNumber}</span>}
                </div>
                <div className="mt-1.5 text-center">
                  <div
                    className={cn('text-xs font-medium', {
                      'text-primary': isCompleted || isCurrent,
                      'text-muted-foreground': isUpcoming,
                    })}
                  >
                    {step.title}
                  </div>
                  {step.description && (
                    <div className="text-xs mt-0.5 text-muted-foreground">{step.description}</div>
                  )}
                </div>
              </div>

              {index < steps.length - 1 && (
                <div
                  className={cn('flex-1 h-px mx-3 transition-colors', {
                    'bg-primary': stepNumber < currentStep,
                    'bg-muted-foreground': stepNumber >= currentStep,
                  })}
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
