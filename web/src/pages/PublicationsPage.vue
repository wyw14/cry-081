<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Search } from 'lucide-vue-next'
import { editorialAPI } from '../services/api'
import type { PublicArticle } from '../types/editorial'
import { useTimezone } from '../composables/useTimezone'

const query = ref('')
const articles = ref<PublicArticle[]>([])
const loading = ref(false)
const total = ref(0)
const { formatUTC } = useTimezone()

async function search() {
  loading.value = true
  try {
    const params = new URLSearchParams({ query: query.value, page: '1', page_size: '20', sort: 'published_at' })
    const page = await editorialAPI.searchPublications(params)
    articles.value = page.Items
    total.value = page.Total
  } finally {
    loading.value = false
  }
}

onMounted(search)
</script>

<template>
  <main class="publications-page">
    <header class="page-header">
      <div><p class="eyebrow">公开刊物库</p><h1>刊发检索</h1></div>
      <span class="result-count">{{ total }} 篇公开文章</span>
    </header>
    <form class="publication-search" @submit.prevent="search">
      <el-input v-model="query" placeholder="按标题、摘要或正文检索"><template #prefix><Search :size="17" /></template></el-input>
      <el-button native-type="submit" type="primary" :loading="loading">检索</el-button>
    </form>
    <section class="publication-list" v-loading="loading">
      <article v-for="article in articles" :key="article.ID">
        <div class="article-meta"><span>{{ article.SectionID }}</span><span>第 {{ article.Edition }} 版</span><time>{{ formatUTC(article.PublishedAt) }}</time></div>
        <h2>{{ article.Title }}</h2>
        <p>{{ article.Abstract }}</p>
        <div class="tag-row"><el-tag v-for="tag in article.Tags" :key="tag" size="small" effect="plain">{{ tag }}</el-tag></div>
      </article>
      <el-empty v-if="!loading && articles.length === 0" description="没有符合条件的公开文章" />
    </section>
  </main>
</template>
