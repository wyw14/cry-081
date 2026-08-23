<script setup lang="ts">
import { computed } from 'vue'
import type { ManuscriptStatus } from '../types/editorial'

const props = defineProps<{ status: ManuscriptStatus }>()

const labels: Record<ManuscriptStatus, string> = {
  draft: '草稿',
  submitted: '待分派',
  under_initial_review: '初审中',
  revision_needed: '待修改',
  under_re_review: '复审中',
  under_final_review: '待终审',
  rejected: '已退稿',
  accepted: '已录用',
  scheduled: '已排期',
  published: '已刊发',
  withdrawn: '已撤回'
}

const tone = computed(() => {
  if (['published', 'accepted'].includes(props.status)) return 'success'
  if (['rejected', 'withdrawn'].includes(props.status)) return 'danger'
  if (['revision_needed', 'under_final_review'].includes(props.status)) return 'warning'
  return 'info'
})
</script>

<template>
  <el-tag :type="tone" size="small" effect="plain">{{ labels[status] }}</el-tag>
</template>
