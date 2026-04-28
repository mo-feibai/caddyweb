import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/web'
    },
    {
        path: '/dashboard',
        redirect: '/web/dashboard'
    },
    {
        path: '/setup',
        name: 'Setup',
        component: () => import('@/views/Setup.vue')
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/Login.vue')
    },
    {
        path: '/web',
        component: () => import('@/views/Layout.vue'),
        children: [
            {
                path: '',
                redirect: 'dashboard'
            },
            {
                path: 'dashboard',
                name: 'Dashboard',
                component: () => import('@/views/Dashboard.vue')
            },
            {
                path: 'domains',
                name: 'Domains',
                component: () => import('@/views/Domains.vue')
            },
            {
                path: 'sites',
                name: 'Sites',
                component: () => import('@/views/Sites.vue')
            },
            {
                path: 'tls',
                name: 'TLS',
                component: () => import('@/views/TLS.vue')
            },
            {
                path: 'logs',
                name: 'Logs',
                component: () => import('@/views/Logs.vue')
            },
            {
                path: 'settings',
                name: 'Settings',
                component: () => import('@/views/Settings.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router
