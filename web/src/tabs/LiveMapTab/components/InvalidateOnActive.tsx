import * as React from 'react'
import { useMap } from 'react-leaflet'
import { IMAGE_BOUNDS } from '../constants'

export const InvalidateOnActive: React.FC = () => {
  const map = useMap()
  React.useEffect(() => {
    const id = setTimeout(() => {
      map.invalidateSize()
      map.fitBounds(IMAGE_BOUNDS)
    }, 50)
    return () => clearTimeout(id)
  }, [map])
  return null
}
