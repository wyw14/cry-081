import { createRouter, createWebHistory } from 'vue-router'
const editorialRoutes = [
  {
    path: '/workbench',
    name: 'editorial-workbench',
    component: () => import('../pages/WorkbenchPage.vue'),
    meta: { workspace: true, title: '稿件工作台' }
  },
  {
    path: '/publications',
    name: 'published-library',
    component: () => import('../pages/PublicationsPage.vue'),
    meta: { workspace: false, title: '刊发检索' }
  },
  {
    path: '/manuscripts/:id',
    name: 'manuscript-detail',
    component: () => import('../pages/WorkbenchPage.vue'),
    meta: { workspace: true, title: '稿件详情' }
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: { name: 'editorial-workbench' } },
    ...editorialRoutes
  ],
  scrollBehavior: () => ({ top: 0 })
})

router.afterEach((route) => {
  document.title = `${String(route.meta.title ?? '墨衡')} · 墨衡`
})
