<script setup lang="ts">
import { BookOpenText, FileSearch, Gauge, Library, Settings, Users } from 'lucide-vue-next'
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark">墨</span><div><strong>墨衡</strong><span>编辑协作平台</span></div></div>
      <nav aria-label="主导航">
        <RouterLink to="/workbench"><Gauge :size="18" />工作台</RouterLink>
        <a href="#"><BookOpenText :size="18" />稿件</a>
        <a href="#"><Users :size="18" />审稿协作</a>
        <RouterLink to="/publications"><Library :size="18" />刊发库</RouterLink>
        <a href="#"><FileSearch :size="18" />审计与报表</a>
      </nav>
      <div class="sidebar-footer"><button title="平台设置"><Settings :size="18" /><span>平台设置</span></button><div class="profile"><span>编</span><div><strong>演示编辑</strong><small>editor@example.test</small></div></div></div>
    </aside>
    <RouterView />
  </div>
</template>
