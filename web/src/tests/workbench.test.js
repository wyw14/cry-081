import { setActivePinia, createPinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useWorkbenchStore } from '../stores/workbench';
vi.mock('../services/api', () => ({
    editorialAPI: {
        listManuscripts: vi.fn().mockResolvedValue({
            Items: [
                { ID: 'm-1', AuthorID: 'a-1', SectionID: 'research', Status: 'submitted', ActiveVersion: 1, Version: 2, UpdatedAt: '2026-08-23T00:00:00Z', Versions: [{ Number: 1, Title: '可复现研究', Abstract: '摘要', Tags: ['research'], Locked: true }] }
            ],
            NextCursor: ''
        })
    }
}));
describe('workbench store', () => {
    beforeEach(() => setActivePinia(createPinia()));
    it('loads manuscripts and derives status counters', async () => {
        const store = useWorkbenchStore();
        await store.load();
        expect(store.manuscripts).toHaveLength(1);
        expect(store.counters.submitted).toBe(1);
        expect(store.error).toBe('');
    });
});
