import NextAuth, { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";
export const authOptions: NextAuthOptions = {
    providers: [
        CredentialsProvider({
            name: "Credentials",
            credentials: {
              username: { label: "Username", type: "text", placeholder: "jsmith" },
              password: { label: "Password", type: "password" }
            },
            // async authorize(credentials, req) {
            //     const {username, password} = credentials as any;
            //     const res = await fetch("http://localhost:8000/auth/login", {
            //         method:"POST",
            //         headers: {
            //             "Content-Type": "application/json",
            //         },
            //         body: JSON.stringify({
            //             username,
            //             password,
            //         }),
            //     });
              
            //   const user = await res.json();
            //     // const res = {
            //     //     ok: true
            //     // };
            //     // const user = {
            //     //     name: "hey"
            //     // };

            //   if (res.ok && user) {
            //     return user;
            //   } else {
            //     return null;
            //   }
            // }
            async authorize(credentials, req) {
                try {
                  //const twinDb = new TwinDB();
                  //const {username, password} = credentials as any;
                  //await twinDb.login(username, password, 'fixed');
                  const user = { id: "id", name: "name", email: "mail" };
                  return user;
                } catch (err) {
                  return null;
                }
                // Add logic here to look up the user from the credentials supplied
                //const user = { id: "1", name: "J Smith", email: "jsmith@example.com" }
                
                //user.name = 'AAA' + twinDb.login('TTT');
                //if (user) {
                  // Any object returned will be saved in `user` property of the JWT
                //  return user
                //} else {
                  // If you return null then an error will be displayed advising the user to check their details.
                //  return null
          
                  // You can also Reject this callback with an Error thus the user will be sent to the error page with the error message as a query parameter
                //}
            }
        })
    ],
    session: {
        strategy: "jwt",
    },
    
    pages: {
        signIn: "/auth/login",
    },
}

export default NextAuth(authOptions);