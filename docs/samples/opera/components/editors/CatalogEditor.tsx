'use client';

import { CatalogState, CatalogSpec } from '@/app/types';
import { useEffect, useState } from 'react';
import { Schema, Rule} from '../../app/types';
import { IoIosAddCircle } from 'react-icons/io';
import { Chip } from "@nextui-org/react";
import Button from '@mui/material/Button';

interface CatalogEditorProps {
    schemas: CatalogState[];
}

function CatalogEditor(props: CatalogEditorProps) {
    const { schemas } = props;
    const [fields, setFields] = useState({});
    const [errors, setErrors] = useState({});
    const [moreFields, setMoreFields] = useState({});
    useEffect(() => {
        if (schemas.length == 0) {
            schemaSelected(schemas[0].spec.name);
        }
    }, [schemas]);
    const handleSchemaChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        schemaSelected(event.target.value);
    };
    const schemaSelected = (value: string) => {
        if (value == "") {
            setFields({});
            setMoreFields({});
            return
        }
        const schema = schemas.find((schema: CatalogState) => schema.spec.name === value);
        if (schema) {
            const spec: Schema = schema.spec.properties['spec'];
            setFields(spec.rules);
            setMoreFields({});
        }
    }
    const handleFormSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        const formData = new FormData(event.currentTarget);
        const data = Object.fromEntries(formData.entries());
        const catalog = {
            name: data.name,
            type: "config"
        };
        if (data.schema) {
            catalog.metadata = {
                "schema": data.schema
            }
        }
        catalog.properties = {};
        Object.keys(data).forEach((key: string) => {
            if (key.includes("-name")) {
                const id = key.split("-")[0];
                const name = data[key];
                const value = data[`${id}-value`];
                if (name && value) {
                    catalog.properties[name] = value;
                }
            } else if (key.includes("-value")) {
                // ignore
            } else if (key != "name" && key != "schema") {
                catalog.properties[key] = data[key];
            } 
        });
        // post to api
        const response = await fetch("/api/catalogs/check", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(catalog),
        });
        const responseData = await response.json();
        setErrors(responseData);
        if (Object.keys(responseData).length == 0) {
            const response = await fetch("/api/catalogs/registry", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(catalog),
            });
            const responseData = await response.json();
            console.log(responseData);
        }
    };

    const addRow = (event: React.MouseEvent<HTMLDivElement, MouseEvent>)  => {        
        event.preventDefault();
        const newFields = {...moreFields};
        const id = `new${Object.keys(moreFields).length}`;
        newFields[id] = {       
            "name": "new",     
        };
        setMoreFields(newFields);
    }
    const removeRow = (event: React.MouseEvent<HTMLDivElement, MouseEvent>)  => {
        event.preventDefault();
        const newFields = {...moreFields};
        const id = event.target.id;
        delete newFields[id];
        setMoreFields(newFields);
    }
    return (
        <div className="container mx-auto max-w-sm">
            <h1 className="text-3xl my-4">Edit Catalog</h1>
            <form className="flex flex-col gap-4 items-stretch" onSubmit={handleFormSubmit}>
                <label htmlFor="name">Name</label>
                <input type="text" id="name" name="name" />
                <label htmlFor="schema">Schema</label>
                <select id="schema" name="schema" onChange={handleSchemaChange}>
                    <option key="empty" value="">---</option>
                    {schemas.map((schema: CatalogState) => <option key={schema.spec.name} value={schema.spec.name}>{schema.spec.name}</option>)}                
                </select>
                <label>Properties</label>
                {fields && Object.keys(fields).map((key: string) => {
                    const field = fields[key];
                    return (
                        <div key={key} className="field_row">
                            <label className="field_name" htmlFor={key}>{key}</label>
                            <div className="field_value">
                                <input  type="text" id={key} name={key} />
                                {errors[key] && <div className="field_error">{errors[key].error}</div>}
                            </div>
                            <div className="field_space"></div>
                        </div>
                    );  
                })}
                {moreFields && Object.keys(moreFields).map((key: string) => {
                    const field = moreFields[key];
                    const nameKey = `${key}-name`;
                    const valueKey = `${key}-value`;
                    const divKey = `${key}-div`;                    
                    return (
                        <div key={divKey} className="field_row">                            
                            <input type="text" id={nameKey} name={nameKey}  className="field_name"/>
                            <div className="field_value">
                                <input type="text" id={valueKey} name={valueKey} />
                                {errors[key] && <div className="field_error">{errors[key].error}</div>}
                            </div>
                            <div className="field_space"><div className="field_button" id={key} onClick={removeRow}>-</div></div>                            
                        </div>
                    );  
                })}   
                <div className="clickable">  
                    <Chip color="success" avatar={<IoIosAddCircle />} onClick={addRow}>Add Row</Chip>                    
                </div>
                <Button type="submit" variant="contained">Submit</Button>
            </form>
            
        </div>
    );
}

export default CatalogEditor;