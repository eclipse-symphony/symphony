/** @type {import('next').NextConfig} */
const nextConfig = {
    images: {
        domains: ['static.boredpanda.com', 'avatars.githubusercontent.com'],
    },
    experimental: {
        serverActions: true,
    },
}

module.exports = nextConfig
