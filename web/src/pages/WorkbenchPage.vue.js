import { onMounted } from 'vue';
import { Filter, RefreshCw, Search } from 'lucide-vue-next';
import SubmissionTable from '../features/submissions/SubmissionTable.vue';
import WorkflowRail from '../components/WorkflowRail.vue';
import { useWorkbenchStore } from '../stores/workbench';
const store = useWorkbenchStore();
onMounted(() => store.load());
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.main, __VLS_intrinsicElements.main)({
    ...{ class: "workbench-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "eyebrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    type: "primary",
}));
const __VLS_2 = __VLS_1({
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "button-icon" },
});
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "summary-band" },
    'aria-label': "工作台摘要",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.store.counters.submitted ?? 0);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
((__VLS_ctx.store.counters.under_initial_review ?? 0) + (__VLS_ctx.store.counters.under_re_review ?? 0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.store.counters.under_final_review ?? 0);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.store.counters.revision_needed ?? 0);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
/** @type {[typeof WorkflowRail, ]} */ ;
// @ts-ignore
const __VLS_4 = __VLS_asFunctionalComponent(WorkflowRail, new WorkflowRail({
    active: (2),
}));
const __VLS_5 = __VLS_4({
    active: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_4));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "table-section" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "toolbar" },
});
const __VLS_7 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
    ...{ class: "search-input" },
    placeholder: "搜索标题、作者或稿件号",
    clearable: true,
}));
const __VLS_9 = __VLS_8({
    ...{ class: "search-input" },
    placeholder: "搜索标题、作者或稿件号",
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_8));
__VLS_10.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_10.slots;
    const __VLS_11 = {}.Search;
    /** @type {[typeof __VLS_components.Search, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        size: (16),
    }));
    const __VLS_13 = __VLS_12({
        size: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
}
var __VLS_10;
const __VLS_15 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.status),
    placeholder: "全部状态",
    clearable: true,
}));
const __VLS_17 = __VLS_16({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.status),
    placeholder: "全部状态",
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_16));
let __VLS_19;
let __VLS_20;
let __VLS_21;
const __VLS_22 = {
    onChange: (...[$event]) => {
        __VLS_ctx.store.load();
    }
};
__VLS_18.slots.default;
const __VLS_23 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
    label: "待分派",
    value: "submitted",
}));
const __VLS_25 = __VLS_24({
    label: "待分派",
    value: "submitted",
}, ...__VLS_functionalComponentArgsRest(__VLS_24));
const __VLS_27 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
    label: "初审中",
    value: "under_initial_review",
}));
const __VLS_29 = __VLS_28({
    label: "初审中",
    value: "under_initial_review",
}, ...__VLS_functionalComponentArgsRest(__VLS_28));
const __VLS_31 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
    label: "待修改",
    value: "revision_needed",
}));
const __VLS_33 = __VLS_32({
    label: "待修改",
    value: "revision_needed",
}, ...__VLS_functionalComponentArgsRest(__VLS_32));
const __VLS_35 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
    label: "待终审",
    value: "under_final_review",
}));
const __VLS_37 = __VLS_36({
    label: "待终审",
    value: "under_final_review",
}, ...__VLS_functionalComponentArgsRest(__VLS_36));
const __VLS_39 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
    label: "已录用",
    value: "accepted",
}));
const __VLS_41 = __VLS_40({
    label: "已录用",
    value: "accepted",
}, ...__VLS_functionalComponentArgsRest(__VLS_40));
var __VLS_18;
const __VLS_43 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
    title: "筛选",
}));
const __VLS_45 = __VLS_44({
    title: "筛选",
}, ...__VLS_functionalComponentArgsRest(__VLS_44));
__VLS_46.slots.default;
const __VLS_47 = {}.Filter;
/** @type {[typeof __VLS_components.Filter, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
    size: (17),
}));
const __VLS_49 = __VLS_48({
    size: (17),
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
var __VLS_46;
const __VLS_51 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    ...{ 'onClick': {} },
    circle: true,
    title: "刷新",
    loading: (__VLS_ctx.store.loading),
}));
const __VLS_53 = __VLS_52({
    ...{ 'onClick': {} },
    circle: true,
    title: "刷新",
    loading: (__VLS_ctx.store.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
let __VLS_55;
let __VLS_56;
let __VLS_57;
const __VLS_58 = {
    onClick: (...[$event]) => {
        __VLS_ctx.store.load();
    }
};
__VLS_54.slots.default;
const __VLS_59 = {}.RefreshCw;
/** @type {[typeof __VLS_components.RefreshCw, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    size: (17),
}));
const __VLS_61 = __VLS_60({
    size: (17),
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
var __VLS_54;
if (__VLS_ctx.store.error) {
    const __VLS_63 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        title: (__VLS_ctx.store.error),
        type: "error",
        showIcon: true,
        closable: (false),
    }));
    const __VLS_65 = __VLS_64({
        title: (__VLS_ctx.store.error),
        type: "error",
        showIcon: true,
        closable: (false),
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
}
/** @type {[typeof SubmissionTable, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(SubmissionTable, new SubmissionTable({
    ...{ 'onOpen': {} },
    rows: (__VLS_ctx.store.manuscripts),
    loading: (__VLS_ctx.store.loading),
}));
const __VLS_68 = __VLS_67({
    ...{ 'onOpen': {} },
    rows: (__VLS_ctx.store.manuscripts),
    loading: (__VLS_ctx.store.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
let __VLS_70;
let __VLS_71;
let __VLS_72;
const __VLS_73 = {
    onOpen: ((id) => __VLS_ctx.$router.push(`/manuscripts/${id}`))
};
var __VLS_69;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "load-more" },
});
if (__VLS_ctx.store.nextCursor) {
    const __VLS_74 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.store.loading),
    }));
    const __VLS_76 = __VLS_75({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.store.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_75));
    let __VLS_78;
    let __VLS_79;
    let __VLS_80;
    const __VLS_81 = {
        onClick: (...[$event]) => {
            if (!(__VLS_ctx.store.nextCursor))
                return;
            __VLS_ctx.store.load(false);
        }
    };
    __VLS_77.slots.default;
    var __VLS_77;
}
/** @type {__VLS_StyleScopedClasses['workbench-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['eyebrow']} */ ;
/** @type {__VLS_StyleScopedClasses['button-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['summary-band']} */ ;
/** @type {__VLS_StyleScopedClasses['table-section']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
/** @type {__VLS_StyleScopedClasses['search-input']} */ ;
/** @type {__VLS_StyleScopedClasses['load-more']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Filter: Filter,
            RefreshCw: RefreshCw,
            Search: Search,
            SubmissionTable: SubmissionTable,
            WorkflowRail: WorkflowRail,
            store: store,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
