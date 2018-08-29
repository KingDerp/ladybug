import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {

  },
  mutations: {

  },
  actions: {
    buyerSignUp({}, buyerSignUpRequest) {
        console.log("entered buyer sign up request")
        console.log(buyerSignUpRequest)
        axios.post('http://localhost:8080/api/buyer/sign-up', buyerSignUpRequest
        )
        .then(response => console.log(response))
        .catch(function (error) {
              console.log("printing error")
              console.log(error);
        });;
    },
    buyerSignIn({}, buyerSignInRequest) {
        console.log("entered buyer sign in request")
        console.log(buyerSignInRequest)
    },
  }
})
