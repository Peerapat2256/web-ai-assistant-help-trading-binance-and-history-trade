export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return;

  const token = useCookie("token");
  const publicPages = ["/login", "/register"];

  if (publicPages.includes(to.path)) {
    if (token.value) {
      return navigateTo("/dashboard");
    }
    return;
  }

  if (!token.value) {
    return navigateTo("/login");
  }
});
