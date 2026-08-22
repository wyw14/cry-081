<script setup lang="ts">
import { ExternalLink, MoreHorizontal } from 'lucide-vue-next'
import { useTimezone } from '../../composables/useTimezone'
import type { Manuscript } from '../../types/editorial'
import StatusBadge from '../../components/StatusBadge.vue'

defineProps<{ rows: Manuscript[]; loading: boolean }>()
const emit = defineEmits<{ open: [id: string] }>()
const { formatUTC } = useTimezone()

function latestTitle(row: Manuscript) {
  return row.Versions.find((version) => version.Number === row.ActiveVersion)?.Title ?? '未命名稿件'
}
</script>

<template>
  <el-table :data="rows" v-loading="loading" row-key="ID" height="460" @row-dblclick="(row: Manuscript) => emit('open', row.ID)">
    <el-table-column label="稿件" min-width="300">
      <template #default="scope">
        <div class="manuscript-cell">
          <button class="title-link" @click="emit('open', scope.row.ID)">{{ latestTitle(scope.row) }}</button>
          <span>{{ scope.row.ID }} · v{{ scope.row.ActiveVersion }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="SectionID" label="栏目" width="130" />
    <el-table-column label="状态" width="120">
      <template #default="scope"><StatusBadge :status="scope.row.Status" /></template>
    </el-table-column>
    <el-table-column label="最近更新" width="180">
      <template #default="scope">{{ formatUTC(scope.row.UpdatedAt) }}</template>
    </el-table-column>
    <el-table-column label="操作" width="112" align="right">
      <template #default="scope">
        <el-button text circle title="打开稿件" @click="emit('open', scope.row.ID)"><ExternalLink :size="17" /></el-button>
        <el-button text circle title="更多操作"><MoreHorizontal :size="18" /></el-button>
      </template>
    </el-table-column>
  </el-table>
</template>
