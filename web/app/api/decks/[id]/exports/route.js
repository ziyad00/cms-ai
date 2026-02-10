import { NextResponse } from 'next/server'

export async function GET(request, { params }) {
  try {
    const { id } = params

    // Forward the request to the Go backend, preserving cookies
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'https://cms-ai-production.up.railway.app'
    const response = await fetch(`${backendUrl}/v1/decks/${id}/exports`, {
      headers: {
        'Accept': 'application/json',
        'Cookie': request.headers.get('cookie') || '',
        'Authorization': request.headers.get('authorization') || '',
      },
    })

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch exports' },
        { status: response.status }
      )
    }

    const data = await response.json()

    // Transform for frontend compatibility
    const transformedExports = data.exports?.map(job => ({
      id: job.id,
      status: job.status === 'Done' || job.status === 'completed' ? 'Done' : job.status,
      type: 'export',
      outputRef: job.outputRef || job.assetId,
      timestamp: job.completedAt || job.updatedAt || job.createdAt,
      progressPct: job.progressPct,
      progressStep: job.progressStep,
      filename: job.metadata?.filename || `export-${job.id.substring(0, 8)}.pptx`
    })) || []

    return NextResponse.json({ exports: transformedExports })
  } catch (error) {
    console.error('Error fetching deck exports:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}