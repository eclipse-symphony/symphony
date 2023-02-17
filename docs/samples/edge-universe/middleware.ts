export { default } from "next-auth/middleware";

export const config = { matcher: ["/me/:path*", "/sources", "/entry", "/"] };