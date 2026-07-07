/**
 * dune-ui — task-specific wrappers over the spice-panel `ui/` component library
 * (Deep Desert Night HUD). Import from here for the higher-level admin widgets
 * (DataTable, PageHeader, SideNav, …); import from `../ui` for base primitives.
 *
 * Side effect: importing this module registers the lucide icon collection
 * with iconify so `<Icon name="..." />` works offline.
 */
import './icons'

export { Icon } from './Icon'
export { PageHeader } from './PageHeader'
export { InfoCard } from './InfoCard'
export { InfoCardItem } from './InfoCardItem'
export { SectionDivider } from './SectionDivider'
export { SectionLabel } from './SectionLabel'
export { Panel } from './Panel'
export { LoadingState } from './LoadingState'
export { DataTable } from './DataTable'
export type { Column } from './DataTable'
export { ActionBar } from '../ui'
export { Dropzone } from './Dropzone'
export { SideNav } from './SideNav'
export { ConfirmDialog } from './ConfirmDialog'
export { NumberInput } from './NumberInput'
export { FieldInput } from './FieldInput'
export { FieldSelect } from './FieldSelect'
export { TimeInput } from './TimeInput'
