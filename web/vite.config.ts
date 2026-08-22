import vuePlugin from '@vitejs/plugin-vue'
import { loadEnv, type ConfigEnv, type UserConfig } from 'vite'

export default function editorialBuild({ mode }: ConfigEnv): UserConfig {
  const settings = loadEnv(mode, process.cwd(), 'EDITORIAL_')
  const upstream = settings.EDITORIAL_BACKEND || 'http://localhost:8080'
  const proxyTargets = ['/api', '/healthz', '/readyz'].reduce<Record<string, string>>(
    (targets, route) => ({ ...targets, [route]: upstream }),
    {}
  )

  return {
    plugins: [vuePlugin()],
    server: { host: true, port: 5173, proxy: proxyTargets },
    build: {
      rollupOptions: {
        output: {
          manualChunks(moduleID) {
            if (moduleID.includes('element-plus')) return 'element'
            if (moduleID.includes('node_modules/vue') || moduleID.includes('pinia')) return 'vue'
          }
        }
      }
    }
  }
}
