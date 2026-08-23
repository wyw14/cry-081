import { computed } from 'vue';
const props = defineProps();
const labels = {
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
};
const tone = computed(() => {
    if (['published', 'accepted'].includes(props.status))
        return 'success';
    if (['rejected', 'withdrawn'].includes(props.status))
        return 'danger';
    if (['revision_needed', 'under_final_review'].includes(props.status))
        return 'warning';
    return 'info';
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    type: (__VLS_ctx.tone),
    size: "small",
    effect: "plain",
}));
const __VLS_2 = __VLS_1({
    type: (__VLS_ctx.tone),
    size: "small",
    effect: "plain",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
(__VLS_ctx.labels[__VLS_ctx.status]);
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            labels: labels,
            tone: tone,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
