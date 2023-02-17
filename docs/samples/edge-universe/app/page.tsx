"use client";
import React, { Suspense } from "react";
import { Tab } from '@headlessui/react';
import { useState, useEffect } from 'react';
import PropertyTile from "@elements/PropertyTile";
import { CSSTransition } from 'react-transition-group';
import { access } from "fs";

function classNames(...classes: string[]) {
    return classes.filter(Boolean).join(' ')
}
function Home() {
    const [perspectivesKeys, setPerspectiveKeys] = useState([]);
    const [perspectives, setPerspectives] = useState([]);

    useEffect(()=> {
    //   const callAPI = async () => {
    //     const ret = new Map();
    //     const response = await fetch('/api/perspectives');
    //     const data = await response.json();        
    //     await Promise.all(data.map(async (p)=>{
    //       const r = await fetch('/api/perspective?name=' + p + '&twin=bob');            
    //       ret.set(p, await r.json());
    //     }));        
    //     return ret;
    //   }
    //   callAPI().then((data)=>{
    //     const arr = Array.from(data.entries());
    //     const keys = arr.map(([key,_])=> key);
    //     const values = arr.map(([_,value])=>value);        
    //     setPerspectiveKeys(keys);
    //     setPerspectives(values);        
    //   });
    },[]);
    
    return (        
        <div className="w-full px-2 py-6 sm:px-5 h-screen">            
            <Tab.Group>
                <Tab.List className="flex space-x-1 rounded-xl bg-blue-900/20 p-1">
                    {perspectivesKeys.map((key, idx) => (
                    <Tab
                        key={idx}
                        className={({ selected }) =>
                            classNames(
                                'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-blue-700',
                                'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2',
                                selected
                                ? 'bg-white shadow'
                                : 'text-blue-500 hover:bg-white/[0.12] hover:text-white'
                            )
                        }
                    >
                    {key}
                    </Tab>
                    ))}
                </Tab.List>
                <Tab.Panels className="mt-2">
                    {perspectives.map((perspective, idx) => (                      
                    <Tab.Panel
                        key={idx}
                        className={classNames(
                        'rounded-xl bg-white p-3',
                        'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2'
                        )}
                    >
                        <CSSTransition timeout={3000} classNames={{
                            enter: 'tile-transition-enter',
                            enterActive: 'tile-transition-enter-active',
                            exit: 'tile-transition-exit',
                            exitActive: 'tile-transition-exit-active'
                          }}>
                          <div className="flex flex-wrap">
                              {Object.entries(perspective.properties).reduce((acc, [key,value]) => {
                                if (key === 'blood_presure.diastolic' || key === 'blood_presure.systolic') {
                                    const group = acc.find(g => g.name === 'blood_presure');
                                    group ? group.values.push({key:key, data: value}): acc.push({name: 'blood_presure', values:[{key: key, data: value}]});
                                } else {
                                    acc.push({name:key, values:[{key:key, data: value}]});
                                }
                                return acc;
                              },[]).map(({name,values}) => {
                                    if (name === 'blood_presure'){
                                        values.sort((a,b) => {
                                            if(a.key === 'blood_presure.diastolic') return 1;
                                            if(b.key === 'blood_presure.diastolic') return -1;
                                            return 0;
                                        });
                                    }
                                    return <PropertyTile key={name} name={name} value={values} mode='current'/>;
                                })}
                          </div>
                        </CSSTransition>                        
                    </Tab.Panel>
                    ))}
                </Tab.Panels>
            </Tab.Group>
        </div>        
    );
}

export default Home;