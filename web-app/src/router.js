import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'
import BuyerSignUp from './views/BuyerSignUp.vue'
import BuyerLogin from './views/BuyerLogin.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/buyer-sign-up',
      name: 'buyer-sign-up',
      component: BuyerSignUp
    },
    {
      path: '/buyer-login',
      name: 'buyer-login',
      component: BuyerLogin
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (about.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import(/* webpackChunkName: "about" */ './views/About.vue')
    }
  ]
})
