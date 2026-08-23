import { computed, ref } from 'vue'

export function useTimezone() {
  const timezone = ref(localStorage.getItem('timezone') ?? 'Asia/Shanghai')

  function formatUTC(value: string | null | undefined) {
    if (!value) return '—'
    return new Intl.DateTimeFormat('zh-CN', {
      timeZone: timezone.value,
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value))
  }

  const label = computed(() => timezone.value.replace('_', ' '))

  function setTimezone(value: string) {
    timezone.value = value
    localStorage.setItem('timezone', value)
  }

  return { timezone, label, formatUTC, setTimezone }
}
