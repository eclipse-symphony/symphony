import { NextResponse } from "next/server"

export async function GET(request: Request) {
    const filteredEnv = Object.entries(process.env)
      .filter(([key]) => !key.startsWith('__NEXT_PRIVATE'))
      .reduce((obj: { [key: string]: string }, [key, value]) => {
        obj[key] = value?.toString() ?? '';
        return obj;
      }, {});
    return NextResponse.json(filteredEnv);
  }