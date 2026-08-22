<script setup lang="ts">
import { onMounted } from 'vue'
import { Filter, RefreshCw, Search } from 'lucide-vue-next'
import SubmissionTable from '../features/submissions/SubmissionTable.vue'
import WorkflowRail from '../components/WorkflowRail.vue'
import { useWorkbenchStore } from '../stores/workbench'

const store = useWorkbenchStore()
onMounted(() => store.load())
</script>

<template>
  <main class="workbench-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">编辑协作台</p>
        <h1>稿件工作台</h1>
      </div>
      <el-button type="primary"><span class="button-icon">+</span> 新建稿件</el-button>
    </header>

    <section class="summary-band" aria-label="工作台摘要">
      <div><strong>{{ store.counters.submitted ?? 0 }}</strong><span>待分派</span></div>
      <div><strong>{{ (store.counters.under_initial_review ?? 0) + (store.counters.under_re_review ?? 0) }}</strong><span>审稿中</span></div>
      <div><strong>{{ store.counters.under_final_review ?? 0 }}</strong><span>待终审</span></div>
      <div><strong>{{ store.counters.revision_needed ?? 0 }}</strong><span>待作者修改</span></div>
      <WorkflowRail :active="2" />
    </section>

    <section class="table-section">
      <div class="toolbar">
        <el-input class="search-input" placeholder="搜索标题、作者或稿件号" clearable>
          <template #prefix><Search :size="16" /></template>
        </el-input>
        <el-select v-model="store.status" placeholder="全部状态" clearable @change="store.load()">
          <el-option label="待分派" value="submitted" />
          <el-option label="初审中" value="under_initial_review" />
          <el-option label="待修改" value="revision_needed" />
          <el-option label="待终审" value="under_final_review" />
          <el-option label="已录用" value="accepted" />
        </el-select>
        <el-button title="筛选"><Filter :size="17" /> 筛选</el-button>
        <el-button circle title="刷新" :loading="store.loading" @click="store.load()"><RefreshCw :size="17" /></el-button>
      </div>
      <el-alert v-if="store.error" :title="store.error" type="error" show-icon :closable="false" />
      <SubmissionTable :rows="store.manuscripts" :loading="store.loading" @open="(id) => $router.push(`/manuscripts/${id}`)" />
      <div class="load-more"><el-button v-if="store.nextCursor" :loading="store.loading" @click="store.load(false)">加载更多</el-button></div>
    </section>
  </main>
</template>
