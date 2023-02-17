import DataSourceList from "./DataSourceList";

export default function RootLayout ({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <main className="flex">
            <div>
                {/* @ts-ignore */}
                <DataSourceList />
            </div>
            <div className="flex-1">{children}</div>
        </main>
    )
}