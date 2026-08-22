import { ExternalLink, MoreHorizontal } from 'lucide-vue-next';
import { useTimezone } from '../../composables/useTimezone';
import StatusBadge from '../../components/StatusBadge.vue';
const __VLS_props = defineProps();
const emit = defineEmits();
const { formatUTC } = useTimezone();
function latestTitle(row) {
    return row.Versions.find((version) => version.Number === row.ActiveVersion)?.Title ?? '未命名稿件';
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onRowDblclick': {} },
    data: (__VLS_ctx.rows),
    rowKey: "ID",
    height: "460",
}));
const __VLS_2 = __VLS_1({
    ...{ 'onRowDblclick': {} },
    data: (__VLS_ctx.rows),
    rowKey: "ID",
    height: "460",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onRowDblclick: ((row) => __VLS_ctx.emit('open', row.ID))
};
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
var __VLS_8 = {};
__VLS_3.slots.default;
const __VLS_9 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
    label: "稿件",
    minWidth: "300",
}));
const __VLS_11 = __VLS_10({
    label: "稿件",
    minWidth: "300",
}, ...__VLS_functionalComponentArgsRest(__VLS_10));
__VLS_12.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_12.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "manuscript-cell" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.emit('open', scope.row.ID);
            } },
        ...{ class: "title-link" },
    });
    (__VLS_ctx.latestTitle(scope.row));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (scope.row.ID);
    (scope.row.ActiveVersion);
}
var __VLS_12;
const __VLS_13 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
    prop: "SectionID",
    label: "栏目",
    width: "130",
}));
const __VLS_15 = __VLS_14({
    prop: "SectionID",
    label: "栏目",
    width: "130",
}, ...__VLS_functionalComponentArgsRest(__VLS_14));
const __VLS_17 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({
    label: "状态",
    width: "120",
}));
const __VLS_19 = __VLS_18({
    label: "状态",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_18));
__VLS_20.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_20.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    /** @type {[typeof StatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(StatusBadge, new StatusBadge({
        status: (scope.row.Status),
    }));
    const __VLS_22 = __VLS_21({
        status: (scope.row.Status),
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
}
var __VLS_20;
const __VLS_24 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "最近更新",
    width: "180",
}));
const __VLS_26 = __VLS_25({
    label: "最近更新",
    width: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_27.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatUTC(scope.row.UpdatedAt));
}
var __VLS_27;
const __VLS_28 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "操作",
    width: "112",
    align: "right",
}));
const __VLS_30 = __VLS_29({
    label: "操作",
    width: "112",
    align: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_31.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_32 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        ...{ 'onClick': {} },
        text: true,
        circle: true,
        title: "打开稿件",
    }));
    const __VLS_34 = __VLS_33({
        ...{ 'onClick': {} },
        text: true,
        circle: true,
        title: "打开稿件",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    let __VLS_36;
    let __VLS_37;
    let __VLS_38;
    const __VLS_39 = {
        onClick: (...[$event]) => {
            __VLS_ctx.emit('open', scope.row.ID);
        }
    };
    __VLS_35.slots.default;
    const __VLS_40 = {}.ExternalLink;
    /** @type {[typeof __VLS_components.ExternalLink, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        size: (17),
    }));
    const __VLS_42 = __VLS_41({
        size: (17),
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    var __VLS_35;
    const __VLS_44 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
        text: true,
        circle: true,
        title: "更多操作",
    }));
    const __VLS_46 = __VLS_45({
        text: true,
        circle: true,
        title: "更多操作",
    }, ...__VLS_functionalComponentArgsRest(__VLS_45));
    __VLS_47.slots.default;
    const __VLS_48 = {}.MoreHorizontal;
    /** @type {[typeof __VLS_components.MoreHorizontal, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        size: (18),
    }));
    const __VLS_50 = __VLS_49({
        size: (18),
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    var __VLS_47;
}
var __VLS_31;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['manuscript-cell']} */ ;
/** @type {__VLS_StyleScopedClasses['title-link']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ExternalLink: ExternalLink,
            MoreHorizontal: MoreHorizontal,
            StatusBadge: StatusBadge,
            emit: emit,
            formatUTC: formatUTC,
            latestTitle: latestTitle,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
