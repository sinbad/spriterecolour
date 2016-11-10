# Sprite Recolour Map Generator

## Introduction

This library is designed to generate the base textures needed to easily recolour 
sprites flexibly on the fly in a shader, while preserving easy authoring of the 
original sprites.

An artist can create a sprite exactly as they usually would, then this tool will 
generate from that a derived "reference" sprite and a base colour parametrisation, which 
re-creates the original appearance. Alternative colour palettes can swapped in
at runtime.

## How to use

```
Usage:
  spriterecolour [options] <input file>

Options:
  -o, --output string    Output sprite file; default <input>_template.png
  -t, --texture string   File to write palette as texture; default <input>_palette.png
```

## Principle

### Previous work

The core idea is an extension of [this blog post](https://gamedevelopment.tutsplus.com/tutorials/how-to-use-a-shader-to-dynamically-swap-a-sprites-colors--cms-25129). In that
case, the Red colour component of the original sprite is used to index a 
texture containing the replacement colours. While this is fine, it requires that
the artist never uses 2 colours with the same Red value, which is overly 
limiting.

### What SpriteRecolour does differently

SpriteRecolour allows the use of any input sprite, and performs these steps

1. Analyse the image and identify all the unique colours
2. Sort the colours into a colour palette ordered by perceptual distance (CIE76)
3. This palette is saved to a palette texture; this is power-of-two sized, and
   at most 256 wide. If more than 256 colours are needed the texture also grows
   vertically in powers of 2.
4. Write another sprite texture the same size as the original, which we call the
   **reference sprite**. Set the colour components as follows:
   * R = U index of colour in palette texture
   * G = V index of colour in palette texture
   * B = Currently unused
   * A = original alpha
5. If the palette is small enough, write out a list of reference palette colours

When rendering the sprite, we simply combine the **reference sprite** with 
*either* a modified palette texture or modified shader arguments to recolour it. 

The recombination algorithm is:

```
out.rgb = GetPaletteColour(in.r, in.g);
out.a = in.a;
```

Where `GetPaletteColour` either samples a palette texture (no filtering /
mipmapping) or indexes an array of shader constants containing the replacement
colour.

## Limitations

### No filtering or mipmaps

Mipmapping or filtering doesn't work because the textures are no longer continuous,
having unrelated values next to each other. Therefore you need to disable all
texture filtering and mipmaps for both the reference sprite and the palette
texture.

### No lossy compression

Because the output reference sprite relies very heavily on correct indexing 
using the Red channel, lossy compression cannot be supported. The reference
sprite is always output in PNG format right now, and you should not convert it
to a lossy compressed format such as JPG or (sadly) ETC/DXT/S3TC. 

#### Example: texture settings in Unity
Click on the sprite texture, and in the Inspector, check the "Override for.."
box and set the format to "Truecolor" (or in Advanced mode, "ARGB 32 bit"). 

You probably also want to make sure that the Packing Tag for the sprite is set
to a value that is only shared by other sprites using recolouring.

Note: if you're using a texture as the recolour palette, make sure you import
this into Unity as a regular texture and that it's not a Sprite, you don't want
to pack it with other sprites. You can use compression for this texture if you
like but watch out for artefacts if adjacent hues are very different.

### Premultiplied alpha

Premultiplied alpha is not supported right now, because the palette RGB and
alpha are stored separately. If you use an input texture with premultiplied
alpha then things may kind of work; but if alpha varies per reference colour
you'll get more colours in the output palette (since say red=200 is 200 at full
alpha and 100 at half-alpha, which looks like different colours to this tool)
so it won't be as easy to supply replacement palettes.
