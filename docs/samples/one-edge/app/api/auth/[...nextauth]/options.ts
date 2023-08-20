import type { NextAuthOptions } from 'next-auth';
import CredentialsProvider from 'next-auth/providers/credentials';

export const options: NextAuthOptions = {
    providers: [
        CredentialsProvider({
            name: "Credentials",
            credentials: {
                username: {
                    label: "Username:",
                    type: "text",
                    placeholder: "your user name"
                },   
                password: {
                    label: "Password:",
                    type: "password",
                    placeholder: "your password"
                },             
            },
            async authorize(credentials) {
                const symphonyApi = process.env.SYMPHONY_API;
  
                const res = await fetch(`${symphonyApi}users/auth`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify(credentials)
                });
                const user = await res.json();
                if (res.ok && user) {                    
                    return user;
                } 
                return null;
            }
        })
    ],
    callbacks: {
        async jwt({token, user}) {
            return {...token, ...user};
        },
        async session({session, token, user}) {
            session.user = token;
            return session;
        }
    }
}