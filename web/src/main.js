import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import App from './App.vue';
import { router } from './router';
import './styles.css';
function mountEditorialDesk(target) {
    const application = createApp(App);
    application.config.errorHandler = (cause) => {
        console.error('editorial interface failed', cause);
    };
    application.provide('editorialTimezone', localStorage.getItem('timezone') ?? 'Asia/Shanghai');
    application.use(createPinia());
    application.use(router);
    application.use(ElementPlus);
    application.mount(target);
}
mountEditorialDesk('#app');
