import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { editorialAPI } from '../services/api';
export const useWorkbenchStore = defineStore('workbench', () => {
    const manuscripts = ref([]);
    const nextCursor = ref('');
    const status = ref('');
    const section = ref('');
    const sort = ref('updated_at');
    const loading = ref(false);
    const error = ref('');
    const counters = computed(() => {
        const values = {};
        for (const manuscript of manuscripts.value) {
            values[manuscript.Status] = (values[manuscript.Status] ?? 0) + 1;
        }
        return values;
    });
    async function load(reset = true) {
        loading.value = true;
        error.value = '';
        try {
            const params = new URLSearchParams({ limit: '25', sort: sort.value });
            if (status.value)
                params.set('status', status.value);
            if (section.value)
                params.set('section', section.value);
            if (!reset && nextCursor.value)
                params.set('cursor', nextCursor.value);
            const page = await editorialAPI.listManuscripts(params);
            manuscripts.value = reset ? page.Items : [...manuscripts.value, ...page.Items];
            nextCursor.value = page.NextCursor;
        }
        catch (cause) {
            error.value = cause instanceof Error ? cause.message : '工作台数据加载失败';
        }
        finally {
            loading.value = false;
        }
    }
    return { manuscripts, nextCursor, status, section, sort, loading, error, counters, load };
});
