"use client";

import {Card, CardHeader, CardBody, CardFooter, Divider, Link} from '@nextui-org/react';
import {FiSettings} from 'react-icons/fi';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell} from "@nextui-org/react";
import { useEffect, useState } from 'react';
import {Chip} from "@nextui-org/react";

interface EnvVars {
  [key: string]: string;
}

export default function Home() {
  const [envVars, setEnvVars] = useState<EnvVars>({});
  const fetchEnvVars = async () => {
    const res = await fetch('/api/env');
    const vars = await res.json();
    setEnvVars(vars);
  };
  useEffect(() => {
    //fetch env vars from server
    fetchEnvVars();    
  }, []);
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">
      <div className="flex flex-col items-center justify-center">
        <Chip variant="light" className="text-4xl font-bold text-center mb-8">{envVars && envVars.APP_TITLE ? envVars.APP_TITLE: "Untitiled"}</Chip>
        {envVars && (
        <Card>
          <CardHeader className="flex gap-3">
            <FiSettings />
            <div className="flex flex-col">
              <p className="text-md font-bold">Environment Variables</p>        
            </div>
          </CardHeader>
          <Divider/>
          <CardBody>
            <Table removeWrapper aria-label="Example static collection table">
              <TableHeader>
                <TableColumn>NAME</TableColumn>
                <TableColumn>VALUE</TableColumn>        
              </TableHeader>
              <TableBody>
                {Object.entries(envVars).map(([key, value]) => (
                  <TableRow key={key}>
                    <TableCell>{key}</TableCell>
                    <TableCell>{value as string}</TableCell>
                  </TableRow>
                ))}        
              </TableBody>
            </Table>
          </CardBody>
          <Divider/>
          <CardFooter>
            <Link isExternal showAnchorIcon href="https://github.com/azure/symphony">
              Visit Symphony source code on GitHub.
            </Link>
          </CardFooter>
        </Card>
        )}
      </div>
    </main>
  )
}
