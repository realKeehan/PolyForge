import { ref } from 'vue'

const copyText = async (text: string): Promise<boolean> => {
  if (!text) {
    return false
  }

  try {
    if (typeof window === 'undefined') {
      return false
    }

    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return true
    }

    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'absolute'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()
    const copied = document.execCommand('copy')
    document.body.removeChild(textarea)
    return copied
  } catch (error) {
    console.error('Failed to copy text', error)
    return false
  }
}

export const useClipboard = (timeout = 2000) => {
  const copied = ref(false)

  const copy = async (text: string) => {
    const success = await copyText(text.trim())

    if (success) {
      copied.value = true
      setTimeout(() => {
        copied.value = false
      }, timeout)
    }

    return success
  }

  return {
    copied,
    copy,
  }
}
