import * as React from 'react'
import { cn } from './lib/cn'

type StepStatus = 'complete' | 'active' | 'upcoming'

const StepperContext = React.createContext<{ currentStep: number }>({ currentStep: 0 })
const StepIndexContext = React.createContext<{ index: number; last: boolean }>({
  index: 0,
  last: false,
})

const useStatus = (): StepStatus => {
  const { currentStep } = React.useContext(StepperContext)
  const { index } = React.useContext(StepIndexContext)
  if (index < currentStep) return 'complete'
  if (index === currentStep) return 'active'
  return 'upcoming'
}

export interface StepperProps extends React.HTMLAttributes<HTMLDivElement> {
  currentStep: number
  size?: 'sm' | 'md'
}

const StepperRoot: React.FC<StepperProps> = ({
  currentStep,
  size: _size,
  className,
  children,
  ...props
}): React.ReactElement => {
  const items = React.Children.toArray(children)
  return (
    <StepperContext.Provider value={{ currentStep }}>
      <div {...props} className={cn('flex items-center', className)}>
        {items.map((child, i) => (
          <StepIndexContext.Provider key={i} value={{ index: i, last: i === items.length - 1 }}>
            {child}
          </StepIndexContext.Provider>
        ))}
      </div>
    </StepperContext.Provider>
  )
}

const Step: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex items-center gap-2.5', className)}>
    {children}
  </div>
)

const Indicator: React.FC = (): React.ReactElement => {
  const status = useStatus()
  const { index } = React.useContext(StepIndexContext)
  return (
    <span
      className={cn(
        'grid size-6 place-items-center font-mono text-xs transition [clip-path:polygon(5px_0,100%_0,100%_calc(100%-5px),calc(100%-5px)_100%,0_100%,0_5px)]',
        status === 'complete' && 'bg-accent text-accent-foreground shadow-[0_0_10px_-2px_var(--accent)]',
        status === 'active' && 'bg-[var(--void)] text-accent shadow-[inset_0_0_0_1px_var(--accent)] hud-glow',
        status === 'upcoming' && 'bg-[var(--void)] text-muted shadow-[inset_0_0_0_1px_var(--steel)]',
      )}
    >
      {status === 'complete' ? '✓' : index + 1}
    </span>
  )
}

const Content: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex flex-col', className)}>
    {children}
  </div>
)

const Title: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => {
  const status = useStatus()
  return (
    <span
      {...props}
      className={cn(
        'text-xs transition',
        status === 'upcoming' ? 'text-muted' : 'text-foreground',
        className,
      )}
    >
      {children}
    </span>
  )
}

const StepperSeparator: React.FC = (): React.ReactElement | null => {
  const status = useStatus()
  const { last } = React.useContext(StepIndexContext)
  if (last) return null
  return (
    <span
      className={cn('mx-3 h-px w-11 transition', status === 'complete' ? 'bg-accent' : 'bg-border')}
    />
  )
}

export const Stepper = Object.assign(StepperRoot, {
  Step,
  Indicator,
  Content,
  Title,
  Separator: StepperSeparator,
})
