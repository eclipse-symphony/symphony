import React from "react";
import { Todo } from "../../../../typings";
import { notFound } from "next/navigation";


export const dynamicParams = true; // default is true, dynaically generate server-side rendered page caches

type PageProps = {
    params: {
        datasourceId: string //corelated with the [datasourceId] name
    }
}

const fetchDatasource = async (dstasourceId: string) => {
    const res = await fetch(`https://jsonplaceholder.typicode.com/todos/${dstasourceId}`, {next: {revalidate: 60}}); // second parameter could be {cache: "no-cache"} for server-side rendering {cache: "force-cache"} for static site generation
    const datasource: Todo = await res.json();
    return datasource;
};

async function DataSourceDetails({params: {datasourceId}} : PageProps) {
    const datasource = await fetchDatasource(datasourceId);

    if (!datasource.id) return notFound();

    return (
        <div className="p-10 bg-yellow-200 border-2 m-2 shadow-lg">
            <p>
                #{datasource.id}: {datasource.title}
            </p>
            <p> Completed: {datasource.completed? "Yes": "No"} </p>
            <p className="border-t border-black mt-5 text-right">
                By User: {datasource.userId}
            </p>
        </div>
    )
}

export default DataSourceDetails;

export async function generateStaticParams() { //a reserved function
    const res = await fetch("https://jsonplaceholder.typicode.com/todos/");
    const todos: Todo[] = await res.json();

    // fo demo, we only prebuild first 10 pages to void rate limit
    const trimmedTodos = todos.splice(0, 10);

    return trimmedTodos.map(todo => ({
        datasourceId: todo.id.toString(),
    }))
}