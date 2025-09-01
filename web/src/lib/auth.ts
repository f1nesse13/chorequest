import { Capacitor } from '@capacitor/core'
import { Preferences } from '@capacitor/preferences'

const TOKEN_KEY = 'auth_token'

export async function setToken(token: string | null) {
  if (!token) {
    if (Capacitor.isNativePlatform()) {
      await Preferences.remove({ key: TOKEN_KEY })
    } else {
      localStorage.removeItem(TOKEN_KEY)
    }
    return
  }
  if (Capacitor.isNativePlatform()) {
    await Preferences.set({ key: TOKEN_KEY, value: token })
  } else {
    localStorage.setItem(TOKEN_KEY, token)
  }
}

export async function getToken(): Promise<string | null> {
  if (Capacitor.isNativePlatform()) {
    const v = await Preferences.get({ key: TOKEN_KEY })
    return v.value ?? null
  }
  return localStorage.getItem(TOKEN_KEY)
}

