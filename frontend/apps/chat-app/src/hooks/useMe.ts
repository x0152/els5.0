import { useEffect, useState } from 'react'
import { api } from '../lib/api'

export type Me = { pictureUrl?: string; initials: string }

export function useMe(): Me | null {
  const [me, setMe] = useState<Me | null>(null)
  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const r = await api.account.accountMe()
        if (!alive || !r) return
        const initials =
          `${r.first_name?.[0] ?? ''}${r.last_name?.[0] ?? ''}`.toUpperCase() ||
          r.email?.[0]?.toUpperCase() ||
          '?'
        setMe({ pictureUrl: r.picture_url || undefined, initials })
      } catch {
        /* ignore */
      }
    })()
    return () => {
      alive = false
    }
  }, [])
  return me
}
