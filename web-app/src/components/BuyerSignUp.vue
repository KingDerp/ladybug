<template>
  <div>
    <input class="display-block-center" v-model="firstName" placeholder="first name">
    <input class="display-block-center" v-model="lastName" placeholder="last name">
    <input class="display-block-center" v-model="password" placeholder="password">
    <input class="display-block-center" v-model="email" placeholder="email">
    <input class="display-block-center" v-model="streetAddress" placeholder="street address">
    <input class="display-block-center" v-model="city" placeholder="city">
    <input class="display-block-center" v-model="state" placeholder="state">
    <input class="display-block-center" v-model="zip" placeholder="zip">
    <button @click="submitSignUp">Submit</button>
  </div>
</template>

<script>

export default {
  name: 'BuyerSignUp',
  data() {
    return {
      firstName: '',
      lastName: '',
      password: '',
      email: '',
      streetAddress: '',
      city: '',
      state: '',
      zip: ''
    }
  },
  methods: {
    submitSignUp() {
      console.log("first: " + this.firstName);
      console.log("last : " + this.lastName);
      console.log("password: " + this.password);
      console.log("email: " + this.email);
      console.log("streetAddress: " + this.streetAddress);
      console.log("city: " + this.city);
      console.log("state: " + this.state);
      console.log("zip: " + this.zip);

      //TODO(mac): pull this and others like it out into a validator function
      if (this.firstName === ""){
        console.log("first name cannot be empty");
        return;
      }

      if (this.lastName === "") {
        console.log("last name cannot be empty");
        return;
      }

      const parsedZip = parseInt(this.zip, 10);
      if (isNaN(parsedZip)) {
        console.log("zip must be an integer");
        return;
      }

      const buyerSignUpRequest = {
        firstName: this.firstName,
        lastName: this.lastName,
        password: this.password,
        email: this.email,
        billingAddress: {
          streetAddress: this.streetAddress,
          city: this.city,
          state: this.state,
          zip: parsedZip
        },
      };

      console.log(buyerSignUpRequest);

      this.$store.dispatch('buyerSignUp', buyerSignUpRequest);
    } 
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">

.display-block-center {
    display: block;
    margin-left: auto;
    margin-right: auto;
}

</style>
