import React from "react";
import { Todo } from '../../../typings';
import Link from "next/link";

const fetchDataSources = async () => {    
    const res = await fetch("https://jsonplaceholder.typicode.com/todos/");
    const todos: Todo[] = await res.json();

    return todos;
}

async function DataSourceList() {
    const dataSources = await fetchDataSources()
    return <>
        {dataSources.map((todo) => (
            <p key={todo.id}>
                <Link href={`me/${todo.id}`}>Todo: {todo.id}</Link>
            </p>
        ))}
    </>;
}

export default DataSourceList;