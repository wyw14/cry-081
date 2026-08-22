<script setup lang="ts">
import { Check, CircleDot } from 'lucide-vue-next'

defineProps<{ active: number }>()

const steps = ['投稿', '编辑初审', '复审', '主编终审', '排期刊发']
</script>

<template>
  <ol class="workflow-rail" aria-label="稿件处理进度">
    <li v-for="(step, index) in steps" :key="step" :class="{ active: index === active, done: index < active }">
      <span class="step-icon">
        <Check v-if="index < active" :size="14" />
        <CircleDot v-else :size="14" />
      </span>
      <span>{{ step }}</span>
    </li>
  </ol>
</template>
